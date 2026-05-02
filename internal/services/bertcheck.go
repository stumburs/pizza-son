package services

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type UserStats struct {
	TotalActivations int            `json:"total"`
	BertCounts       map[string]int `json:"bert_counts"`
}

type ChannelData struct {
	Berts     []string             `json:"berts"`
	UserStats map[string]UserStats `json:"user_stats"`
}

type BertService struct {
	mu   sync.RWMutex
	Data map[string]*ChannelData // channel name -> data
	path string
}

type BertStats struct {
	TotalBertchecks        int // times bertchecked
	MostCommonBert         string
	MostCommonCount        int
	BertsCollectedOutOfAll int // collectec out of all berts count
	TotalBerts             int // all bert count
}

var BertServiceInstance *BertService

func NewBertService() {
	service := &BertService{
		Data: make(map[string]*ChannelData),
		path: "data/bertcheck/stats.json",
	}
	service.load()
	BertServiceInstance = service
	log.Println("[Bert] Service initialized")

}

func (s *BertService) GetBerts(channel string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if d, ok := s.Data[channel]; ok {
		return d.Berts
	}
	return []string{}
}

func (s *BertService) RegisterActivation(channel, user, bert string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureChannel(channel)

	stats := s.Data[channel].UserStats[user]
	if stats.BertCounts == nil {
		stats.BertCounts = make(map[string]int)
	}

	stats.TotalActivations++
	stats.BertCounts[bert]++
	s.Data[channel].UserStats[user] = stats
	s.save()
}

func (s *BertService) AddBert(channel, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureChannel(channel)

	// prevent duplicates
	if slices.Contains(s.Data[channel].Berts, name) {
		return
	}

	s.Data[channel].Berts = append(s.Data[channel].Berts, name)
	s.save()
}

func (s *BertService) RemoveBert(channel, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureChannel(channel)

	originalLen := len(s.Data[channel].Berts)
	newBerts := []string{}
	for _, b := range s.Data[channel].Berts {
		if b != name {
			newBerts = append(newBerts, b)
		}
	}

	s.Data[channel].Berts = newBerts
	s.save()
	return len(newBerts) < originalLen
}

func (s *BertService) GetUserStats(channel, user string) BertStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.Data[channel]
	if !ok {
		return BertStats{}
	}
	stats, ok := data.UserStats[user]
	if !ok {
		return BertStats{
			TotalBerts: len(data.Berts),
		}
	}

	bestBert, max := "", -1
	for name, count := range stats.BertCounts {
		if count > max {
			max = count
			bestBert = name
		}
	}
	return BertStats{
		TotalBertchecks:        stats.TotalActivations,
		MostCommonBert:         bestBert,
		MostCommonCount:        max,
		BertsCollectedOutOfAll: len(stats.BertCounts),
		TotalBerts:             len(data.Berts),
	}
}

func (s *BertService) ensureChannel(channel string) {
	if _, ok := s.Data[channel]; !ok {
		s.Data[channel] = &ChannelData{
			Berts:     []string{},
			UserStats: make(map[string]UserStats),
		}
	}
}

func (s *BertService) save() {
	_ = os.MkdirAll(filepath.Dir(s.path), 0755)
	b, _ := json.MarshalIndent(s.Data, "", "  ")
	_ = os.WriteFile(s.path, b, 0644)
}

func (s *BertService) load() {
	b, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(b, &s.Data)
	}
}
