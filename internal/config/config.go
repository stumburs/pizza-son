package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Twitch struct {
		User  string `toml:"user"`
		OAuth string `toml:"oauth"`
	} `toml:"twitch"`

	Bot struct {
		Prefix   string   `toml:"prefix"`
		Channels []string `toml:"channels"`
	} `toml:"bot"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
