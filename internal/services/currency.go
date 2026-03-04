package services

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
)

const currencyFile = "data/currency/balances.json"

type CurrencyService struct {
	mu       sync.Mutex
	balances map[string]int // key - user ID
}

type UserBalance struct {
	UserID  string
	Balance int
}

var CurrencyServiceInstance *CurrencyService

func NewCurrencyService() {
	if err := os.MkdirAll("data/currency", os.ModePerm); err != nil {
		log.Fatal("[Currency] Failed to create directory:", err)
	}

	svc := &CurrencyService{
		balances: make(map[string]int),
	}

	svc.load()
	CurrencyServiceInstance = svc
	log.Println("[Currency] Service initialized")
}

func (s *CurrencyService) load() {
	data, err := os.ReadFile(currencyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("[Currency] Failed to load balances:", err)
	}
	if err := json.Unmarshal(data, &s.balances); err != nil {
		log.Println("[Currency] Failed to parse balances:", err)
	}
	log.Printf("[Currency] Loaded %d balances", len(s.balances))
}

func (s *CurrencyService) save() {
	data, err := json.MarshalIndent(s.balances, "", "  ")
	if err != nil {
		log.Println("[Currency] Failed to marshal balances:", err)
		return
	}
	if err := os.WriteFile(currencyFile, data, 0644); err != nil {
		log.Println("[Currency] Failed to save balances:", err)
	}
}

func (s *CurrencyService) Balance(userID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.balances[userID]
}

func (s *CurrencyService) Add(userID string, amount int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] += amount
	s.save()
	return s.balances[userID]
}

// Returns false if failed to deduct
func (s *CurrencyService) Deduct(userID string, amount int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.balances[userID] < amount {
		return s.balances[userID], false
	}
	s.balances[userID] -= amount
	if s.balances[userID] < 0 {
		s.balances[userID] = 0
	}
	s.save()
	return s.balances[userID], true
}

func (s *CurrencyService) Set(userID string, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] = amount
	s.save()
}

func (s *CurrencyService) Give(fromID, toID string, amount int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.balances[fromID] < amount {
		return s.balances[fromID], false
	}
	s.balances[fromID] -= amount
	s.balances[toID] += amount
	s.save()
	return s.balances[fromID], true
}

func (s *CurrencyService) TopBalances(n int) []UserBalance {
	s.mu.Lock()
	defer s.mu.Unlock()

	all := make([]UserBalance, 0, len(s.balances))
	for id, bal := range s.balances {
		all = append(all, UserBalance{UserID: id, Balance: bal})
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Balance > all[j].Balance
	})

	if n > len(all) {
		n = len(all)
	}

	return all[:n]
}
