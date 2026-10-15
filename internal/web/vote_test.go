package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func setupVoteStore(t *testing.T) *VoteStore {
	t.Helper()
	store := NewVoteStore(filepath.Join(t.TempDir(), "votes.json"))
	// Make sure GetVoteStore never overwrites the test instance afterwards.
	voteStoreOnce.Do(func() {})
	old := voteStoreInstance
	voteStoreInstance = store
	t.Cleanup(func() { voteStoreInstance = old })
	return store
}

func TestVoteBordaTally(t *testing.T) {
	store := setupVoteStore(t)

	if _, err := store.AddBallot("alice", []string{"a", "b", "c", "d", "e"}); err != nil {
		t.Fatalf("AddBallot: %v", err)
	}
	if _, err := store.AddBallot("bob", []string{"e", "d", "c", "b", "a"}); err != nil {
		t.Fatalf("AddBallot: %v", err)
	}

	res := store.Results()
	if res.TotalBallots != 2 {
		t.Errorf("total ballots = %d, want 2", res.TotalBallots)
	}
	// Symmetric ballots: every bert should end up with 4 points.
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		if res.Totals[name] != 4 {
			t.Errorf("totals[%s] = %d, want 4", name, res.Totals[name])
		}
		if res.Appearances[name] != 2 {
			t.Errorf("appearances[%s] = %d, want 2", name, res.Appearances[name])
		}
	}
}

func TestVoteValidation(t *testing.T) {
	store := setupVoteStore(t)

	if _, err := store.AddBallot("x", []string{"only-one"}); err == nil {
		t.Error("expected error for single-item ranking")
	}
	if _, err := store.AddBallot("x", []string{"a", "a", "b"}); err == nil {
		t.Error("expected error for duplicate names")
	}
	if _, err := store.AddBallot("x", []string{"", "  "}); err == nil {
		t.Error("expected error for empty ranking")
	}
	if _, err := store.AddBallot("", []string{"a", "b", "c"}); err == nil {
		t.Error("expected error for missing voter name")
	}
	if _, err := store.AddBallot("   ", []string{"a", "b", "c"}); err == nil {
		t.Error("expected error for blank voter name")
	}
	if got := store.Results().TotalBallots; got != 0 {
		t.Errorf("invalid ballots stored: total = %d", got)
	}
}

func TestVotePersistenceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "votes.json")
	store := NewVoteStore(path)
	if _, err := store.AddBallot("alice", []string{"a", "b", "c"}); err != nil {
		t.Fatalf("AddBallot: %v", err)
	}

	reloaded := NewVoteStore(path)
	res := reloaded.Results()
	if res.TotalBallots != 1 {
		t.Fatalf("reloaded ballots = %d, want 1", res.TotalBallots)
	}
	if res.Totals["a"] != 2 || res.Totals["b"] != 1 || res.Totals["c"] != 0 {
		t.Errorf("wrong reloaded totals: %v", res.Totals)
	}
}

func TestVoteHTTPHandlers(t *testing.T) {
	setupVoteStore(t)
	ws := &WebService{}

	body, _ := json.Marshal(map[string]any{
		"voter":   "alice",
		"ranking": []string{"a", "b", "c", "d", "e"},
	})
	rec := httptest.NewRecorder()
	ws.handleVoteSubmit(rec, httptest.NewRequest(http.MethodPost, "/api/vote/submit", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("submit status = %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	ws.handleVoteSubmit(rec, httptest.NewRequest(http.MethodPost, "/api/vote/submit", bytes.NewReader([]byte("{bad"))))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad body status = %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	ws.handleVoteResults(rec, httptest.NewRequest(http.MethodGet, "/api/vote/results", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("results status = %d", rec.Code)
	}
	var res voteResults
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode results: %v", err)
	}
	if res.TotalBallots != 1 || res.Totals["a"] != 4 || res.Totals["e"] != 0 {
		t.Errorf("wrong HTTP results: %+v", res)
	}
}
