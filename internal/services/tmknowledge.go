package services

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/store"
	"strings"
	"sync"
	"time"
)

type TMKnowledgeEntry struct {
	Content string `json:"content"`
	AddedBy string `json:"added_by"`
	AddedAt string `json:"added_at"`
}

type TMKnowledgeData struct {
	Entries []TMKnowledgeEntry `json:"entries"`
}

type TMKnowledgeService struct {
	mu    sync.Mutex
	store *store.Store[TMKnowledgeData]
}

var TMKnowledgeServiceInstance *TMKnowledgeService

func NewTMKnowledgeService() {
	data := &TMKnowledgeData{}
	s := &TMKnowledgeService{
		store: store.New("data/tmknowledge.json", data),
	}
	s.store.EnsureDir()

	_, err := s.store.LoadIfExists()
	if err != nil {
		fmt.Println("[TMKnowledge] Failed to load:", err)
	}

	TMKnowledgeServiceInstance = s
	fmt.Println("[TMKnowledge] Loaded", len(data.Entries), "entries")
}

func (s *TMKnowledgeService) Add(content, addedBy string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := s.store.Data()
	entry := TMKnowledgeEntry{
		Content: content,
		AddedBy: addedBy,
		AddedAt: time.Now().Format("2006-01-02"),
	}
	d.Entries = append(d.Entries, entry)
	s.store.Save()
	return len(d.Entries)
}

func (s *TMKnowledgeService) Remove(number int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := s.store.Data()
	if number < 1 || number > len(d.Entries) {
		return false
	}
	d.Entries = append(d.Entries[:number-1], d.Entries[number+1:]...)
	s.store.Save()
	return true
}

func (s *TMKnowledgeService) Random() (TMKnowledgeEntry, int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := s.store.Data()
	if len(d.Entries) == 0 {
		return TMKnowledgeEntry{}, 0, false
	}
	idx := rand.IntN(len(d.Entries))
	return d.Entries[idx], idx + 1, true
}

func (s *TMKnowledgeService) Get(number int) (TMKnowledgeEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := s.store.Data()
	if number < 1 || number > len(d.Entries) {
		return TMKnowledgeEntry{}, false
	}
	return d.Entries[number-1], true
}

func (s *TMKnowledgeService) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.store.Data().Entries)
}

func (s *TMKnowledgeService) GetKnowledgeText() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := s.store.Data()
	if len(d.Entries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Trackmania knowledge:\n")
	for _, e := range d.Entries {
		b.WriteString(fmt.Sprintf("- %s\n", e.Content))
	}
	return b.String()
}
