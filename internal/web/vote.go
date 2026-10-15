package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// VoteBallot is one ranked round, ordered best -> worst.
type VoteBallot struct {
	Voter     string    `json:"voter"`
	Ranking   []string  `json:"ranking"`
	Timestamp time.Time `json:"timestamp"`
}

// VoteStore persists ballots to a local JSON file so results survive restarts.
type VoteStore struct {
	mu      sync.Mutex
	path    string
	ballots []VoteBallot
}

var voteStoreInstance *VoteStore
var voteStoreOnce sync.Once

func GetVoteStore() *VoteStore {
	voteStoreOnce.Do(func() {
		voteStoreInstance = NewVoteStore(filepath.Join("data", "votes.json"))
	})
	return voteStoreInstance
}

func NewVoteStore(path string) *VoteStore {
	s := &VoteStore{path: path}
	if data, err := os.ReadFile(path); err == nil {
		var ballots []VoteBallot
		if err := json.Unmarshal(data, &ballots); err == nil {
			s.ballots = ballots
		}
	}
	return s
}

// AddBallot validates and stores a ballot. Ranking must be best -> worst,
// contain at least 2 distinct non-empty names. Voter is required.
func (s *VoteStore) AddBallot(voter string, ranking []string) (VoteBallot, error) {
	voter = strings.TrimSpace(voter)
	if voter == "" {
		return VoteBallot{}, errInvalidVote("voter name is required")
	}
	if len(voter) > 40 {
		return VoteBallot{}, errInvalidVote("voter name too long")
	}
	cleaned := make([]string, 0, len(ranking))
	seen := make(map[string]bool)
	for _, name := range ranking {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if seen[name] {
			return VoteBallot{}, errDuplicateVoteName(name)
		}
		seen[name] = true
		cleaned = append(cleaned, name)
	}
	if len(cleaned) < 2 {
		return VoteBallot{}, errInvalidVote("ranking must contain at least 2 berts")
	}

	ballot := VoteBallot{
		Voter:     voter,
		Ranking:   cleaned,
		Timestamp: time.Now().UTC(),
	}

	s.mu.Lock()
	s.ballots = append(s.ballots, ballot)
	toSave := make([]VoteBallot, len(s.ballots))
	copy(toSave, s.ballots)
	s.mu.Unlock()

	if err := s.save(toSave); err != nil {
		return VoteBallot{}, err
	}
	return ballot, nil
}

func (s *VoteStore) save(ballots []VoteBallot) error {
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if ballots == nil {
		ballots = []VoteBallot{}
	}
	data, err := json.MarshalIndent(ballots, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

type voteResults struct {
	TotalBallots int            `json:"total_ballots"`
	Totals       map[string]int `json:"totals"`
	Appearances  map[string]int `json:"appearances"`
	Voters       []string       `json:"voters"`
}

// Results tallies Borda points: in a ballot of n names, position i
// (0 = best) scores n-1-i points.
func (s *VoteStore) Results() voteResults {
	s.mu.Lock()
	defer s.mu.Unlock()

	totals := make(map[string]int)
	appearances := make(map[string]int)
	voterSet := make(map[string]bool)
	for _, b := range s.ballots {
		n := len(b.Ranking)
		for i, name := range b.Ranking {
			totals[name] += n - 1 - i
			appearances[name]++
		}
		if strings.TrimSpace(b.Voter) != "" {
			voterSet[strings.TrimSpace(b.Voter)] = true
		}
	}
	voters := make([]string, 0, len(voterSet))
	for v := range voterSet {
		voters = append(voters, v)
	}
	sort.Strings(voters)

	return voteResults{
		TotalBallots: len(s.ballots),
		Totals:       totals,
		Appearances:  appearances,
		Voters:       voters,
	}
}

func (ws *WebService) handleVoteSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Voter   string   `json:"voter"`
		Ranking []string `json:"ranking"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Ranking) > 32 {
		http.Error(w, "Ranking too large", http.StatusBadRequest)
		return
	}

	ballot, err := GetVoteStore().AddBallot(req.Voter, req.Ranking)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ballot)
}

func (ws *WebService) handleVoteResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetVoteStore().Results())
}

type voteError string

func (e voteError) Error() string { return string(e) }

func errInvalidVote(msg string) error {
	return voteError(msg)
}

func errDuplicateVoteName(name string) error {
	return voteError("duplicate bert in ranking: " + name)
}
