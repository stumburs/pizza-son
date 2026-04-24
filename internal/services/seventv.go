package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const sevenTVDir = "data/7tv"
const sevenTVAPIURL = "https://7tv.io/v3/users/twitch/%s"

type SevenTVEmote struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type SevenTVService struct {
	mu     sync.RWMutex
	emotes map[string][]SevenTVEmote // channel -> emote
}

var SevenTVServiceInstance *SevenTVService

func NewSevenTVService() {
	if err := os.MkdirAll(sevenTVDir, os.ModePerm); err != nil {
		log.Fatal("[7TV] Failed to create directory:", err)
	}
	svc := &SevenTVService{
		emotes: make(map[string][]SevenTVEmote),
	}

	// Load any existing emote files
	svc.loadAll()
	SevenTVServiceInstance = svc
	log.Println("[7TV] Service initialized")
}

func (s *SevenTVService) loadAll() {
	entries, err := os.ReadDir(sevenTVDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		channel := strings.TrimSuffix(entry.Name(), ".json")
		emotes, err := s.loadFromFile(channel)
		if err != nil {
			log.Printf("[7TV] Failed to load emotes for %s: %v", channel, err)
			continue
		}
		s.emotes[channel] = emotes
		log.Printf("[7TV] Loaded %d emotes for %s", len(emotes), channel)
	}
}

func (s *SevenTVService) loadFromFile(channel string) ([]SevenTVEmote, error) {
	path := fmt.Sprintf("%s/%s.json", sevenTVDir, channel)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var emotes []SevenTVEmote
	return emotes, json.Unmarshal(data, &emotes)
}

func (s *SevenTVService) saveToFile(channel string, emotes []SevenTVEmote) error {
	path := fmt.Sprintf("%s/%s.json", sevenTVDir, channel)
	data, err := json.Marshal(emotes)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Fetches emotes from 7TV api and saves them to file
func (s *SevenTVService) Fetch(channel string) (int, error) {
	twitchID, err := TwitchServiceInstance.GetUserID(channel)
	if err != nil {
		return 0, fmt.Errorf("could not resolve Twitch ID for %s: %w", channel, err)
	}

	url := fmt.Sprintf(sevenTVAPIURL, twitchID)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return 0, fmt.Errorf("%s has no 7TV emote set", channel)
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("7TV API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		EmoteSet struct {
			Emotes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"emotes"`
		} `json:"emote_set"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	emotes := make([]SevenTVEmote, 0, len(result.EmoteSet.Emotes))
	for _, e := range result.EmoteSet.Emotes {
		if e.Name != "" && e.ID != "" {
			emotes = append(emotes, SevenTVEmote{Name: e.Name, ID: e.ID})
		}
	}

	s.mu.Lock()
	s.emotes[channel] = emotes
	s.mu.Unlock()

	if err := s.saveToFile(channel, emotes); err != nil {
		return 0, fmt.Errorf("failed to save emotes: %w", err)
	}

	return len(emotes), nil
}

func (s *SevenTVService) GetEmotes(channel string) []SevenTVEmote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.emotes[channel]
}

func (s *SevenTVService) HasEmote(channel, emote string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.emotes[channel] {
		if e.Name == emote {
			return true
		}
	}
	return false
}
