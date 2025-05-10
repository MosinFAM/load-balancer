package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port     int      `json:"port"`
	Backends []string `json:"backends"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}
	return &cfg, nil
}
