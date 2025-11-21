package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() Config {
	return Config{
		TimeFormat:            "24h",
		LocationDisplayFormat: "auto",
		Colleagues: []Colleague{
			{
				Name:      "Alice (New York)",
				Timezone:  "America/New_York",
				WorkStart: 9,
				WorkEnd:   17,
			},
			{
				Name:      "Bob (London)",
				Timezone:  "Europe/London",
				WorkStart: 9,
				WorkEnd:   17,
			},
			{
				Name:      "Charlie (Tokyo)",
				Timezone:  "Asia/Tokyo",
				WorkStart: 9,
				WorkEnd:   17,
			},
		},
	}
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "tui-clock", "config.yaml"), nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, create it with defaults
			config := DefaultConfig()
			if err := SaveConfig(path, config); err != nil {
				return config, fmt.Errorf("failed to create default config: %w", err)
			}
			return config, nil
		}
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for colleagues if not specified
	for i := range config.Colleagues {
		if config.Colleagues[i].WorkStart == 0 {
			config.Colleagues[i].WorkStart = 9
		}
		if config.Colleagues[i].WorkEnd == 0 {
			config.Colleagues[i].WorkEnd = 17
		}
	}

	// Set default time format if not specified
	if config.TimeFormat == "" {
		config.TimeFormat = "24h"
	}

	// Set default location display format if not specified
	if config.LocationDisplayFormat == "" {
		config.LocationDisplayFormat = "auto"
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(path string, config Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
