package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Auth struct {
		Twitch struct {
			Username string `toml:"username"`
			OAuth    string `toml:"oauth"`
		} `toml:"twitch"`
	} `toml:"auth"`
}

// Load config and return parsed data
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
