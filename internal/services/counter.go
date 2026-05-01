package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

const counterFile = "data/counters.json"

type Counter struct {
	Name    string `json:"name"`
	Value   int    `json:"value"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type CounterService struct {
	mu       sync.Mutex
	counters map[string]map[string]*Counter // channel -> name -> counter
}

var CounterServiceInstance *CounterService

func NewCounterService() {
	svc := &CounterService{
		counters: make(map[string]map[string]*Counter),
	}
	svc.load()
	CounterServiceInstance = svc
	log.Println("[Counter] Service initialized")
}

func (s *CounterService) load() {
	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("[Counter] Failed to load:", err)
		return
	}
	var flat []*Counter
	if err := json.Unmarshal(data, &flat); err != nil {
		log.Println("[Counter] Failed to parse:", err)
		return
	}
	for _, c := range flat {
		if _, ok := s.counters[c.Channel]; !ok {
			s.counters[c.Channel] = make(map[string]*Counter)
		}
		s.counters[c.Channel][c.Name] = c
	}
	log.Printf("[Counter] Loaded %d counters", len(flat))
}

func (s *CounterService) save() {
	flat := make([]*Counter, 0)
	for _, counters := range s.counters {
		for _, c := range counters {
			flat = append(flat, c)
		}
	}
	data, err := json.MarshalIndent(flat, "", "  ")
	if err != nil {
		log.Println("[Counter] Failed to marshal:", err)
		return
	}
	os.WriteFile(counterFile, data, 0644)
}

func (s *CounterService) Add(channel, name, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.counters[channel]; !ok {
		s.counters[channel] = make(map[string]*Counter)
	}
	if _, exists := s.counters[channel][name]; exists {
		return fmt.Errorf("counter %q already exists", name)
	}
	if message == "" {
		message = fmt.Sprintf("%s: {}", name)
	}
	s.counters[channel][name] = &Counter{Name: name, Value: 0, Channel: channel, Message: message}
	s.save()
	return nil
}

func (s *CounterService) Remove(channel, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.counters[channel][name]; !ok {
		return fmt.Errorf("counter %q not found", name)
	}
	delete(s.counters[channel], name)
	s.save()
	return nil
}

func (s *CounterService) Increment(channel, name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.counters[channel][name]
	if !ok {
		return 0, fmt.Errorf("counter %q not found", name)
	}
	c.Value++
	s.save()
	return c.Value, nil
}

func (s *CounterService) Set(channel, name string, value int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.counters[channel][name]
	if !ok {
		return fmt.Errorf("counter %q not found", name)
	}
	c.Value = value
	s.save()
	return nil
}

func (s *CounterService) Get(channel, name string) (*Counter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.counters[channel][name]
	return c, ok
}

func (s *CounterService) List(channel string) []*Counter {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*Counter, 0)
	for _, c := range s.counters[channel] {
		result = append(result, c)
	}
	return result
}

func (s *CounterService) Exists(channel, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.counters[channel][name]
	return ok
}
