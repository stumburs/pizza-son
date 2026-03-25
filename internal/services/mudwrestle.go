package services

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const mudwrestleFile = "data/mudwrestle/mudwrestle.json"

type MudwrestleStats struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	SlicesWon  int `json:"slices_won"`
	SlicesLost int `json:"slices_lost"`
}

type MudwrestleService struct {
	mu    sync.Mutex
	stats map[string]*MudwrestleStats // key: userID
}

var MudwrestleServiceInstance *MudwrestleService

func NewMudwrestleService() {
	if err := os.MkdirAll("data/mudwrestle", os.ModePerm); err != nil {
		log.Fatal("[Mudwrestle] Failed to create directory:", err)
	}

	svc := &MudwrestleService{
		stats: make(map[string]*MudwrestleStats),
	}
	svc.load()
	MudwrestleServiceInstance = svc
	log.Println("[Mudwrestle] Service initialized")
}

func (s *MudwrestleService) load() {
	data, err := os.ReadFile(mudwrestleFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("[Mudwrestle] Failed to load:", err)
		return
	}
	if err := json.Unmarshal(data, &s.stats); err != nil {
		log.Println("[Mudwrestle] Failed to parse:", err)
	}
}

func (s *MudwrestleService) save() {
	data, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		log.Println("[Mudwrestle] Failed to marshal:", err)
		return
	}
	os.WriteFile(mudwrestleFile, data, 0644)
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
