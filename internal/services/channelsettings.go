package services

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const channelSettingsFile = "data/channel_settings.json"

type ChannelSettings struct {
	DisabledCommands map[string]bool `json:"disabled_commands"`
}

type ChannelSettingsService struct {
	mu       sync.Mutex
	settings map[string]*ChannelSettings // key - channel name
}

var ChannelSettingsInstance *ChannelSettingsService

func NewChannelSettingsService() {
	svc := &ChannelSettingsService{
		settings: make(map[string]*ChannelSettings),
	}

	svc.load()
	ChannelSettingsInstance = svc
	log.Println("[ChannelSettings] Service initialized")
}

func (s *ChannelSettingsService) load() {
	data, err := os.ReadFile(channelSettingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("[ChannelSettings] Failed to load:", err)
		return
	}
	if err := json.Unmarshal(data, &s.settings); err != nil {
		log.Println("[ChannelSettings] Failed to parse:", err)
	}
	log.Printf("[ChannelSettings] Loaded settings for %d channels", len(s.settings))
}

func (s *ChannelSettingsService) save() {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		log.Println("[ChannelSettings] Failed to marshal:", err)
		return
	}
	if err := os.WriteFile(channelSettingsFile, data, 0644); err != nil {
		log.Println("[ChannelSettings] Failed to save:", err)
	}
}

func (s *ChannelSettingsService) getOrCreate(channel string) *ChannelSettings {
	if _, ok := s.settings[channel]; !ok {
		s.settings[channel] = &ChannelSettings{
			DisabledCommands: make(map[string]bool),
		}
	}
	return s.settings[channel]
}

func (s *ChannelSettingsService) IsCommandEnabled(channel, command string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cs, ok := s.settings[channel]
	if !ok {
		return true
	}
	return !cs.DisabledCommands[command]
}

func (s *ChannelSettingsService) EnableCommand(channel, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cs := s.getOrCreate(channel)
	delete(cs.DisabledCommands, command)
	s.save()
}

func (s *ChannelSettingsService) DisableCommand(channel, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cs := s.getOrCreate(channel)
	cs.DisabledCommands[command] = true
	s.save()
}

func (s *ChannelSettingsService) ListDisabled(channel string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	cs, ok := s.settings[channel]
	if !ok {
		return []string{}
	}
	disabled := make([]string, 0)
	for cmd := range cs.DisabledCommands {
		disabled = append(disabled, cmd)
	}
	return disabled
}
