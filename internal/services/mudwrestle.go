package services

import (
	"log"
	"pizza-son/internal/store"
	"sync"
)

type MudwrestleStats struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	SlicesWon  int `json:"slices_won"`
	SlicesLost int `json:"slices_lost"`
}

type MudwrestleService struct {
	mu    sync.Mutex
	store *store.Store[map[string]*MudwrestleStats]
	stats map[string]*MudwrestleStats // key: userID
}

var MudwrestleServiceInstance *MudwrestleService

func NewMudwrestleService() {
	s := store.New("data/mudwrestle/mudwrestle.json", &map[string]*MudwrestleStats{})
	s.EnsureDir()
	svc := &MudwrestleService{
		store: s,
		stats: *s.Data(),
	}
	svc.load()
	MudwrestleServiceInstance = svc
	log.Println("[Mudwrestle] Service initialized")
}

func (s *MudwrestleService) load() {
	if _, err := s.store.LoadIfExists(); err != nil {
		log.Println("[Mudwrestle] Failed to load:", err)
	}
}

func (s *MudwrestleService) save() {
	if err := s.store.Save(); err != nil {
		log.Println("[Mudwrestle] Failed to save:", err)
	}
}

func (s *MudwrestleService) getOrCreate(userID string) *MudwrestleStats {
	if _, ok := s.stats[userID]; !ok {
		s.stats[userID] = &MudwrestleStats{}
	}
	return s.stats[userID]
}

func (s *MudwrestleService) RecordWin(userID string, slices int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.getOrCreate(userID)
	st.Wins++
	st.SlicesWon += slices
	s.save()
}

func (s *MudwrestleService) RecordLoss(userID string, slices int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.getOrCreate(userID)
	st.Losses++
	st.SlicesLost += slices
	s.save()
}

func (s *MudwrestleService) GetStats(userID string) MudwrestleStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.stats[userID]; ok {
		return *st
	}
	return MudwrestleStats{}
}
