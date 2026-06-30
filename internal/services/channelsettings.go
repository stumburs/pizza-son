package services

import (
	"log"
	"pizza-son/internal/store"
	"slices"
	"strings"
	"sync"
)

type ChannelSettings struct {
	DisabledCommands  map[string]bool `json:"disabled_commands"`
	DisabledListeners map[string]bool `json:"disabled_listeners"`
}

type ChannelSettingsService struct {
	mu       sync.RWMutex
	store    *store.Store[map[string]*ChannelSettings]
	settings map[string]*ChannelSettings // key - channel name
}

var ChannelSettingsInstance *ChannelSettingsService

func NewChannelSettingsService() {
	s := store.New("data/channel_settings.json", &map[string]*ChannelSettings{})
	svc := &ChannelSettingsService{
		store:    s,
		settings: *s.Data(),
	}
	svc.load()
	ChannelSettingsInstance = svc
	log.Println("[ChannelSettings] Service initialized")
}

func (s *ChannelSettingsService) load() {
	ok, err := s.store.LoadIfExists()
	if err != nil {
		log.Println("[ChannelSettings] Failed to load:", err)
		return
	}
	if ok {
		log.Printf("[ChannelSettings] Loaded settings for %d channels", len(s.settings))
	}
}

func (s *ChannelSettingsService) save() {
	if err := s.store.Save(); err != nil {
		log.Println("[ChannelSettings] Failed to save:", err)
	}
}

func (s *ChannelSettingsService) getOrCreate(channel string) *ChannelSettings {
	if _, ok := s.settings[channel]; !ok {
		s.settings[channel] = &ChannelSettings{
			DisabledCommands:  make(map[string]bool),
			DisabledListeners: make(map[string]bool),
		}
	}
	cs := s.settings[channel]
	if cs.DisabledCommands == nil {
		cs.DisabledCommands = make(map[string]bool)
	}
	if cs.DisabledListeners == nil {
		cs.DisabledListeners = make(map[string]bool)
	}
	return cs
}

// Commands
func (s *ChannelSettingsService) IsCommandEnabled(channel, command string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, ok := s.settings[channel]
	if !ok {
		return true
	}
	return !cs.DisabledCommands[strings.ToLower(command)]
}

func (s *ChannelSettingsService) EnableCommand(channel, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.getOrCreate(channel).DisabledCommands, strings.ToLower(command))
	s.save()
}

func (s *ChannelSettingsService) DisableCommand(channel, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getOrCreate(channel).DisabledCommands[strings.ToLower(command)] = true
	s.save()
}

func (s *ChannelSettingsService) ListDisabledCommands(channel string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, ok := s.settings[channel]
	if !ok {
		return []string{}
	}
	return listDisabled(cs.DisabledCommands)
}

// Listeners
func (s *ChannelSettingsService) IsListenerEnabled(channel, listener string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, ok := s.settings[channel]
	if !ok {
		return true
	}
	return !cs.DisabledListeners[strings.ToLower(listener)]
}

func (s *ChannelSettingsService) EnableListener(channel, listener string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.getOrCreate(channel).DisabledListeners, strings.ToLower(listener))
	s.save()
}

func (s *ChannelSettingsService) DisableListener(channel, listener string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getOrCreate(channel).DisabledListeners[strings.ToLower(listener)] = true
	s.save()
}

func (s *ChannelSettingsService) ListDisabledListeners(channel string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, ok := s.settings[channel]
	if !ok {
		return []string{}
	}
	return listDisabled(cs.DisabledListeners)
}

// Shared
func listDisabled(m map[string]bool) []string {
	if m == nil {
		return []string{}
	}
	disabled := make([]string, 0, len(m))
	for name, isDisabled := range m {
		if isDisabled {
			disabled = append(disabled, name)
		}
	}
	slices.Sort(disabled)
	return disabled
}
