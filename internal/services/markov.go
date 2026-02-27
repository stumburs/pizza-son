package services

import (
	"fmt"
	"log"
	"os"
	"pizza-son/internal/config"
	"sync"

	"github.com/stumburs/mgo"
)

const (
	ngramsDir = "markov_data"
)

type MarkovService struct {
	mu            sync.Mutex
	generators    map[string]*mgo.MarkovGenerator
	messageCounts map[string]int
}

var MarkovServiceInstance *MarkovService

func NewMarkovService() {
	if err := os.MkdirAll(ngramsDir, os.ModePerm); err != nil {
		log.Println("[Markov] Failed to create ngrams directory:", err)
	}

	MarkovServiceInstance = &MarkovService{
		generators:    make(map[string]*mgo.MarkovGenerator),
		messageCounts: make(map[string]int),
	}
}

func ngramsPath(channel string) string {
	return fmt.Sprintf("%s/%s.bin", ngramsDir, channel)
}

func (s *MarkovService) getOrCreate(channel string) *mgo.MarkovGenerator {
	if g, ok := s.generators[channel]; ok {
		return g
	}
	g := mgo.NewMarkovGenerator()

	path := ngramsPath(channel)
	if _, err := os.Stat(path); err == nil {
		g.ReadNgrams(path)
		log.Printf("[Markov] Loaded ngrams for channel %s from %s", channel, path)
	}

	s.generators[channel] = g
	return g
}

func (s *MarkovService) Learn(channel, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g := s.getOrCreate(channel)
	g.SourceText = text
	g.BuildNgrams(mgo.SplitByNCharacters, 4)
	s.messageCounts[channel]++

	if s.messageCounts[channel]%config.Get().Markov.AutosaveInterval == 0 {
		path := ngramsPath(channel)
		g.WriteNgrams(path)
		log.Printf("[Markov] Saved ngrams for channel %s to %s (%d messages)", channel, path, s.messageCounts[channel])
	}
}

func (s *MarkovService) Generate(channel string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageCounts[channel] < 1 {
		return ""
	}

	g := s.getOrCreate(channel)
	return g.GenerateText(config.Get().Markov.LengthToGenerate)
}
