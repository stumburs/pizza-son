package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const tokensPath = "data/tokens.toml"

type tokens struct {
	AccessToken  string `toml:"access_token"`
	RefreshToken string `toml:"refresh_token"`
}

var errTokensMissing = errors.New("tokens missing")

func loadTokens(path string) (tokens, error) {
	var t tokens
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := saveTokens(path, t); err != nil {
			return t, err
		}
		return t, errTokensMissing
	}

	if err != nil {
		return t, err
	}

	if err := toml.Unmarshal(data, &t); err != nil {
		return t, err
	}
	return t, nil
}

func saveTokens(path string, t tokens) error {
	data, err := toml.Marshal(t)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func refreshTokens(clientID, clientSecret, refreshToken string) (tokens, error) {
	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
	if err != nil {
		return tokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tokens{}, fmt.Errorf("refresh failed: %s", resp.Status)
	}

	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return tokens{}, err
	}

	return tokens{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}
