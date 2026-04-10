package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// No racist jokes
const jokeAPIURL = "https://v2.jokeapi.dev/joke/Any?blacklistFlags=racist&type=single"

type JokeResponse struct {
	Error    bool   `json:"error"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Joke     string `json:"joke"`
	Flags    Flags  `json:"flags"`
	ID       int    `json:"id"`
	Safe     bool   `json:"safe"`
	Lang     string `json:"lang"`
}

type Flags struct {
	NSFW      bool `json:"nsfw"`
	Religious bool `json:"religious"`
	Political bool `json:"political"`
	Racist    bool `json:"racist"`
	Sexist    bool `json:"sexist"`
	Explicit  bool `json:"explicit"`
}

func GetJoke() (JokeResponse, error) {
	resp, err := http.Get(jokeAPIURL)
	if err != nil {
		return JokeResponse{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return JokeResponse{}, fmt.Errorf("Joke API returned status %d", resp.StatusCode)
	}

	var jokeResp JokeResponse
	if err := json.NewDecoder(resp.Body).Decode(&jokeResp); err != nil {
		return JokeResponse{}, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if jokeResp.Error {
		return JokeResponse{}, fmt.Errorf("API returned error")
	}

	return jokeResp, nil
}
