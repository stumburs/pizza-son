package services

import (
	"log"
	"sync"
)

type PredictionOutcome struct {
	ID    string
	Title string
}

type PredictionBet struct {
	UserID    string
	OutcomeID string
	Amount    int
}

type Prediction struct {
	ID       string
	Title    string
	Outcomes []PredictionOutcome
	Bets     []PredictionBet
	Active   bool
}

type PredictionService struct {
	mu          sync.Mutex
	predictions map[string]*Prediction // key: channel
}

var PredictionServiceInstance *PredictionService

func NewPredictionService() {
	PredictionServiceInstance = &PredictionService{
		predictions: make(map[string]*Prediction),
	}
	log.Println("[Prediction] Service initialized")
}

func (s *PredictionService) Start(channel, id, title string, outcomes []PredictionOutcome) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.predictions[channel] = &Prediction{
		ID:       id,
		Title:    title,
		Outcomes: outcomes,
		Bets:     []PredictionBet{},
		Active:   true,
	}
	log.Printf("[Prediction] Started in %s: %s", channel, title)
}

func (s *PredictionService) End(channel, winningOutcomeID string) []PredictionBet {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.predictions[channel]
	if !ok || !p.Active {
		return nil
	}
	p.Active = false

	// Total pot and winning pot
	totalPot := 0
	winningPot := 0
	for _, bet := range p.Bets {
		totalPot += bet.Amount
		if bet.OutcomeID == winningOutcomeID {
			winningPot += bet.Amount
		}
	}

	winners := make([]PredictionBet, 0)

	if winningPot == 0 {
		// Refund everyone
		for _, bet := range p.Bets {
			CurrencyServiceInstance.Add(bet.UserID, bet.Amount)
		}
		delete(s.predictions, channel)
		return winners
	}

	// Proportional payout for winners
	for _, bet := range p.Bets {
		if bet.OutcomeID == winningOutcomeID {
			payout := int(float64(bet.Amount) / float64(winningPot) * float64(totalPot))
			CurrencyServiceInstance.Add(bet.UserID, payout)
			winners = append(winners, bet)
		}
	}

	delete(s.predictions, channel)
	return winners
}

func (s *PredictionService) Cancel(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.predictions[channel]
	if !ok {
		return
	}
	// Refund
	for _, bet := range p.Bets {
		CurrencyServiceInstance.Add(bet.UserID, bet.Amount)
		delete(s.predictions, channel)
		log.Printf("[Prediction] Cancelled in %s, all bets refunded", channel)
	}
}

func (s *PredictionService) PlaceBet(channel, userID, outcomeID string, amount int) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.predictions[channel]
	if !ok || !p.Active {
		return "No active prediction in this channel.", false
	}

	// Check valid outcome
	validOutcome := false
	for _, o := range p.Outcomes {
		if o.ID == outcomeID {
			validOutcome = true
			break
		}
	}
	if !validOutcome {
		return "Invalid outcome.", false
	}

	// Check user hasn't already bet
	for _, bet := range p.Bets {
		if bet.UserID == userID {
			return "You have already placed a bet.", false
		}
	}

	p.Bets = append(p.Bets, PredictionBet{
		UserID:    userID,
		OutcomeID: outcomeID,
		Amount:    amount,
	})
	return "", true
}

func (s *PredictionService) GetActive(channel string) (*Prediction, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.predictions[channel]
	if !ok || !p.Active {
		return nil, false
	}
	return p, true
}
