package services

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type BertRecord struct {
	Count     int       `json:"count"`
	ZazaCount int       `json:"zaza_count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type UserStats struct {
	TotalActivations  int                   `json:"total"`
	TotalZazas        int                   `json:"total_zazas"`
	BertRecords       map[string]BertRecord `json:"bert_records"`
	DailyActivations  map[string]int        `json:"daily_activations"`
	HourlyActivations map[string]int        `json:"hourly_activations"`
	BertCounts        map[string]int        `json:"bert_counts,omitempty"` // legacy support for migrating
}

type ChannelData struct {
	Berts             []string             `json:"berts"`
	UserStats         map[string]UserStats `json:"user_stats"`
	DailyActivations  map[string]int       `json:"daily_activations"`
	HourlyActivations map[string]int       `json:"hourly_activations"`
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

func (s *BertService) RegisterActivation(channel, user, bert string, isZaza bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.ensureChannel(channel)

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	hourStr := now.Format("2006-01-02 15:00")

	chData := s.Data[channel]

	// log channel timeline
	if chData.DailyActivations == nil {
		chData.DailyActivations = make(map[string]int)
	}
	chData.DailyActivations[dateStr]++

	if chData.HourlyActivations == nil {
		chData.HourlyActivations = make(map[string]int)
	}
	chData.HourlyActivations[hourStr]++

	stats := chData.UserStats[user]
	if stats.BertRecords == nil {
		stats.BertRecords = make(map[string]BertRecord)
	}
	if stats.DailyActivations == nil {
		stats.DailyActivations = make(map[string]int)
	}
	if stats.HourlyActivations == nil {
		stats.HourlyActivations = make(map[string]int)
	}

	// fetch or initialize item record
	record := stats.BertRecords[bert]
	if record.Count == 0 {
		record.FirstSeen = now
	}
	record.Count++
	record.LastSeen = now

	if isZaza {
		record.ZazaCount++
		stats.TotalZazas++
	}

	// save updates back to state
	stats.BertRecords[bert] = record
	stats.TotalActivations++
	stats.DailyActivations[dateStr]++
	stats.HourlyActivations[hourStr]++

	chData.UserStats[user] = stats

	// look up emote ID and broadcast
	var emoteID string
	if SevenTVServiceInstance != nil {
		emoteID = SevenTVServiceInstance.GetEmoteID(channel, bert)
	}
	LiveFeedInstance.Broadcast(user, channel, bert, emoteID)

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
		for bert, record := range uStats.BertRecords {
			if activeBerts[bert] {
				channelTotal += record.Count
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
	for bert, record := range stats.BertRecords {
		if !activeBerts[bert] {
			continue
		}
		activeTotal += record.Count
		collectedBerts++
		if record.Count > max {
			max = record.Count
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
			Berts:             []string{},
			UserStats:         make(map[string]UserStats),
			DailyActivations:  make(map[string]int),
			HourlyActivations: make(map[string]int),
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
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &s.Data)

	// automated migrating
	migrated := false
	now := time.Now()

	for chName, chData := range s.Data {
		for username, stats := range chData.UserStats {
			// check if contains legacy counter maps
			if len(stats.BertCounts) > 0 {
				if stats.BertRecords == nil {
					stats.BertRecords = make(map[string]BertRecord)
				}

				for legacyBert, count := range stats.BertCounts {
					stats.BertRecords[legacyBert] = BertRecord{
						Count:     count,
						ZazaCount: 0,
						FirstSeen: now,
						LastSeen:  now,
					}
				}

				// nullify legacy fields
				stats.BertCounts = nil
				chData.UserStats[username] = stats
				migrated = true
			}
		}
		s.Data[chName] = chData
	}

	if migrated {
		log.Println("[Bert] Legacy database records successfully transformed to new timeline format!")
		s.save()
	}
}
