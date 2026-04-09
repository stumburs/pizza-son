package services

import (
	"bufio"
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

type SevenTVService struct {
	mu     sync.RWMutex
	emotes map[string][]string // channel -> emote names
}

var SevenTVServiceInstance *SevenTVService

func NewSevenTVService() {
	if err := os.MkdirAll(sevenTVDir, os.ModePerm); err != nil {
		log.Fatal("[7TV] Failed to create directory:", err)
	}
	svc := &SevenTVService{
		emotes: make(map[string][]string),
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		channel := strings.TrimSuffix(entry.Name(), ".txt")
		emotes, err := s.loadFromFile(channel)
		if err != nil {
			log.Printf("[7TV] Failed to load emotes for %s: %v", channel, err)
			continue
		}
		s.emotes[channel] = emotes
		log.Printf("[7TV] Loaded %d emotes for %s", len(emotes), channel)
	}
}

func (s *SevenTVService) loadFromFile(channel string) ([]string, error) {
	path := fmt.Sprintf("%s/%s.txt", sevenTVDir, channel)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var emotes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			emotes = append(emotes, line)
		}
	}
	return emotes, scanner.Err()
}

func (s *SevenTVService) saveToFile(channel string, emotes []string) error {
	path := fmt.Sprintf("%s/%s.txt", sevenTVDir, channel)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, e := range emotes {
		fmt.Fprintln(f, e)
	}
	return nil
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
				Name string `json:"name"`
			} `json:"emotes"`
		} `json:"emote_set"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	emotes := make([]string, 0, len(result.EmoteSet.Emotes))
	for _, e := range result.EmoteSet.Emotes {
		if e.Name != "" {
			emotes = append(emotes, e.Name)
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

func (s *SevenTVService) GetEmotes(channel string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.emotes[channel]
}

func (s *SevenTVService) HasEmote(channel, emote string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.emotes[channel] {
		if e == emote {
			return true
		}
	}
	return false
}
