package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	AutoUpdateEnabled bool `json:"auto_update_enabled"`
	UpdateInterval    int  `json:"update_interval_seconds"`
}

func DefaultConfig() *Config {
	return &Config{
		AutoUpdateEnabled: true,
		UpdateInterval:    60,
	}
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), err
	}

	configPath := filepath.Join(homeDir, ".claude-code-monitor", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".claude-code-monitor")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
