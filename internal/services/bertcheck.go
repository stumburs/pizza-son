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
	Mu   sync.RWMutex
	Data map[string]*ChannelData // channel name -> data
	path string
}

type BertStats struct {
	TotalBertchecks        int // times bertchecked
	MostCommonBert         string
	MostCommonCount        int
	BertsCollectedOutOfAll int // collected out of all berts count
	TotalBerts             int // all bert count
	ChannelTotalBertchecks int // total bertchecks done by all chatters
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
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	if d, ok := s.Data[channel]; ok {
		return d.Berts
	}
	return []string{}
}

func (s *BertService) RegisterActivation(channel, user, bert string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
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
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.ensureChannel(channel)

	// prevent duplicates
	if slices.Contains(s.Data[channel].Berts, name) {
		return
	}

	s.Data[channel].Berts = append(s.Data[channel].Berts, name)
	s.save()
}

func (s *BertService) RemoveBert(channel, name string) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
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
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	data, ok := s.Data[channel]
	if !ok {
		return BertStats{}
	}

	activeBerts := make(map[string]bool, len(data.Berts))
	for _, b := range data.Berts {
		activeBerts[b] = true
	}

	// channel total only counts currently active berts
	channelTotal := 0
	for _, uStats := range data.UserStats {
		for bert, count := range uStats.BertCounts {
			if activeBerts[bert] {
				channelTotal += count
			}
		}
	}

	stats, ok := data.UserStats[user]
	if !ok {
		return BertStats{
			TotalBerts:             len(data.Berts),
			ChannelTotalBertchecks: channelTotal,
		}
	}

	// only count activations of currently active berts
	activeTotal := 0
	bestBert, max := "", -1
	collectedBerts := 0
	for bert, count := range stats.BertCounts {
		if !activeBerts[bert] {
			continue
		}
		activeTotal += count
		collectedBerts++
		if count > max {
			max = count
			bestBert = bert
		}
	}

	return BertStats{
		TotalBertchecks:        activeTotal,
		MostCommonBert:         bestBert,
		MostCommonCount:        max,
		BertsCollectedOutOfAll: collectedBerts,
		TotalBerts:             len(data.Berts),
		ChannelTotalBertchecks: channelTotal,
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
