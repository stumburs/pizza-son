package config

import (
	"log"
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

	Ollama struct {
		Host       string `toml:"host"`
		Model      string `toml:"model"`
		NumPredict int    `toml:"num_predict"`
		MaxHistory int    `toml:"max_history"`
	} `toml:"ollama"`
}

var cfg *Config

func Load(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	cfg = &c
}

func Get() *Config {
	if cfg == nil {
		log.Fatal("Config not loaded. Call config.Load() first.")
	}
	return cfg
}
