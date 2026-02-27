package services

import (
	"sync"

	"github.com/stumburs/mgo"
)

type MarkovService struct {
	mu            sync.Mutex
	generators    map[string]*mgo.MarkovGenerator
	messageCounts map[string]int
}

var MarkovServiceInstance *MarkovService

func NewMarkovService() {
	MarkovServiceInstance = &MarkovService{
		generators:    make(map[string]*mgo.MarkovGenerator),
		messageCounts: make(map[string]int),
	}
}

func (s *MarkovService) getOrCreate(channel string) *mgo.MarkovGenerator {
	if g, ok := s.generators[channel]; ok {
		return g
	}
	g := mgo.NewMarkovGenerator()
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
}

func (s *MarkovService) Generate(channel string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messageCounts[channel] < 1 {
		return ""
	}

	g := s.getOrCreate(channel)
	return g.GenerateText(100)
}
