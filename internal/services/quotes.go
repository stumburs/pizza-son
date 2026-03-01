package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"sync"
	"time"
)

const quotesDir = "quotes"

type Quote struct {
	Text      string `json:"text"`
	AddedBy   string `json:"added_by"`
	CreatedAt string `json:"created_at"`
}

type QuoteService struct {
	mu     sync.Mutex
	quotes map[string][]Quote
}

var QuoteServiceInstance *QuoteService

func NewQuoteService() {
	if err := os.MkdirAll(quotesDir, os.ModePerm); err != nil {
		log.Println("[Quotes] Failed to create quotes directory:", err)
	}
	QuoteServiceInstance = &QuoteService{
		quotes: make(map[string][]Quote),
	}
	log.Println("[Quotes] Service initialized")
}

func quotesPath(channel string) string {
	return fmt.Sprintf("%s/%s.json", quotesDir, channel)
}

func (s *QuoteService) getOrLoad(channel string) []Quote {
	if q, ok := s.quotes[channel]; ok {
		return q
	}

	path := quotesPath(channel)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		s.quotes[channel] = []Quote{}
		return s.quotes[channel]
	}

	f, err := os.Open(path)
	if err != nil {
		log.Printf("[Quotes] Failed to open quotes for %s: %v", channel, err)
		return []Quote{}
	}
	defer f.Close()

	var quotes []Quote
	if err := json.NewDecoder(f).Decode(&quotes); err != nil {
		log.Printf("[Quotes] Failed to decode quotes for %s: %v", channel, err)
		return []Quote{}
	}

	s.quotes[channel] = quotes
	log.Printf("[Quotes] Loaded %d quotes for channel %s", len(quotes), channel)
	return quotes
}

func (s *QuoteService) save(channel string) {
	f, err := os.Create(quotesPath(channel))
	if err != nil {
		log.Printf("[Quotes] Failed to save quotes for %s: %v", channel, err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(s.quotes[channel])
}

func (s *QuoteService) Add(channel, text, addedBy string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	quotes := s.getOrLoad(channel)
	quote := Quote{
		Text:      text,
		AddedBy:   addedBy,
		CreatedAt: time.Now().Format("2006-01-02"),
	}
	s.quotes[channel] = append(quotes, quote)
	s.save(channel)
	return len(s.quotes[channel])
}

func (s *QuoteService) Random(channel string) (Quote, int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quotes := s.getOrLoad(channel)
	if len(quotes) == 0 {
		return Quote{}, 0, false
	}
	idx := rand.IntN(len(quotes))
	return quotes[idx], idx + 1, true
}

func (s *QuoteService) Get(channel string, number int) (Quote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quotes := s.getOrLoad(channel)
	if number < 1 || number > len(quotes) {
		return Quote{}, false
	}
	return quotes[number-1], true
}

func (s *QuoteService) Count(channel string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.getOrLoad(channel))
}
