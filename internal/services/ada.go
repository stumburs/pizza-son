package services

import (
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"strings"
	"sync"
)

const adaDataDir = "ada_data"

var defaultDatabase = []ConversationPair{
	{"Hi!", "Hello! How are you?"},
	{"Hello!", "Hi there!"},
	{"How are you?", "I'm fine, thanks for asking!"},
	{"What's your name?", "I'm Ada, your lovely chatbot."},
	{"What do you like?", "I like chatting with you!"},
}

type ConversationPair struct {
	Input    string `json:"input"`
	Response string `json:"response"`
}

type AdaService struct {
	mu             sync.Mutex
	databases      map[string][]ConversationPair
	lastBotMessage map[string]string
}

var AdaServiceInstance *AdaService

func NewAdaService() {
	if err := os.MkdirAll(adaDataDir, os.ModePerm); err != nil {
		log.Println("[Ada] Failed to create data directory:", err)
	}

	AdaServiceInstance = &AdaService{
		databases:      make(map[string][]ConversationPair),
		lastBotMessage: make(map[string]string),
	}

	log.Println("[Ada] Now active and learning!")
}

func adaPath(channel string) string {
	return adaDataDir + "/" + channel + ".json"
}

func (s *AdaService) getOrLoadDatabase(channel string) []ConversationPair {
	if db, ok := s.databases[channel]; ok {
		return db
	}

	path := adaPath(channel)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		db := make([]ConversationPair, len(defaultDatabase))
		copy(db, defaultDatabase)
		s.databases[channel] = db
		log.Printf("[Ada] New database for channel %s", channel)
		return db
	}

	f, err := os.Open(path)
	if err != nil {
		log.Printf("[Ada] Failed to open database for %s: %v", channel, err)
		return defaultDatabase
	}
	defer f.Close()

	var db []ConversationPair
	if err := json.NewDecoder(f).Decode(&db); err != nil {
		log.Printf("[Ada] Failed to decode database for %s: %v", channel, err)
		return defaultDatabase
	}

	s.databases[channel] = db
	log.Printf("[Ada] Loaded %d pairs for channel %s", len(db), channel)
	return db
}

func (s *AdaService) saveDatabase(channel string) {
	db, ok := s.databases[channel]
	if !ok {
		return
	}
	f, err := os.Create(adaPath(channel))
	if err != nil {
		log.Printf("[Ada] Failed to save database for %s: %v", channel, err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(db)
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	return strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z' || '0' <= r && r <= '9')
	})
}

func tfIdfVectorize(docs []string) []map[string]float64 {
	tokenized := make([][]string, len(docs))
	for i, doc := range docs {
		tokenized[i] = tokenize(doc)
	}

	N := float64(len(docs))
	df := make(map[string]float64)
	for _, tokens := range tokenized {
		seen := make(map[string]bool)
		for _, t := range tokens {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}

	vectors := make([]map[string]float64, len(docs))
	for i, tokens := range tokenized {
		tf := make(map[string]float64)
		for _, t := range tokens {
			tf[t]++
		}
		vec := make(map[string]float64)
		for term, count := range tf {
			idf := math.Log((N + 1) / (df[term] + 1))
			vec[term] = (count / float64(len(tokens))) * idf
		}
		vectors[i] = vec
	}
	return vectors
}

func consineSimilarity(a, b map[string]float64) float64 {
	dot, normA, normB := 0.0, 0.0, 0.0
	for term, va := range a {
		dot += va * b[term]
		normA += va * va
	}
	for _, vb := range b {
		normB += vb * vb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (s *AdaService) findBestResponse(db []ConversationPair, userMessage string, topN int) (string, float64) {
	inputs := make([]string, len(db))
	for i, pair := range db {
		inputs[i] = pair.Input
	}

	vectors := tfIdfVectorize(append(inputs, userMessage))
	userVec := vectors[len(vectors)-1]
	inputsVecs := vectors[:len(vectors)-1]

	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, len(inputsVecs))
	for i, vec := range inputsVecs {
		scores[i] = scored{i, consineSimilarity(userVec, vec)}
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	if len(scores) > topN {
		scores = scores[:topN]
	}

	if scores[0].score < 0.2 {
		return "", scores[0].score
	}

	total := 0.0
	for _, sc := range scores {
		total += sc.score
	}
	r := rand.Float64() * total
	cumulative := 0.0
	for _, sc := range scores {
		cumulative += sc.score
		if r <= cumulative {
			return db[sc.idx].Response, sc.score
		}
	}

	return db[scores[0].idx].Response, scores[0].score
}

func (s *AdaService) GetResponse(channel, userInput string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return ""
	}

	db := s.getOrLoadDatabase(channel)

	if last, ok := s.lastBotMessage[channel]; ok && last != "" {
		s.databases[channel] = append(s.databases[channel], ConversationPair{
			Input:    last,
			Response: userInput,
		})

		go func(ch string) {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.saveDatabase(ch)
		}(channel)
		db = s.databases[channel]
	}

	reply, score := s.findBestResponse(db, userInput, 3)
	log.Printf("[Ada] [%s] %q -> %q (score: %.3f)", channel, userInput, reply, score)

	if reply == "" {
		fallbacks := []string{"Interesting!", "Tell me more.", "Hmm, okay.", "Why do you say that?"}
		reply = fallbacks[rand.IntN(len(fallbacks))]
	}
	s.lastBotMessage[channel] = reply
	return reply
}
