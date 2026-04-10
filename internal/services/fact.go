package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// https://uselessfacts.jsph.pl
const randomFactAPI = "https://uselessfacts.jsph.pl/api/v2/facts/random"

type FactResponse struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Source    string `json:"source"`
	SourceURL string `json:"source_url"`
	Language  string `json:"language"`
	Permalink string `json:"permalink"`
}

func GetFact() (FactResponse, error) {
	resp, err := http.Get(randomFactAPI)
	if err != nil {
		return FactResponse{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FactResponse{}, fmt.Errorf("Joke API returned status %d", resp.StatusCode)
	}

	var factResp FactResponse
	if err := json.NewDecoder(resp.Body).Decode(&factResp); err != nil {
		return FactResponse{}, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return factResp, nil
}
