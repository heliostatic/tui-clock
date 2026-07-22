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
		ColorScheme:           "classic",
		TimelineMode:          "individual",
		Colleagues: []Colleague{
			{
				Name:     "Alice (New York)",
				Timezone: "America/New_York",
			},
			{
				Name:     "Bob (London)",
				Timezone: "Europe/London",
			},
			{
				Name:     "Charlie (Tokyo)",
				Timezone: "Asia/Tokyo",
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

	return parseConfig(data)
}

// parseConfig unmarshals config data and normalizes it (defaults and
// legacy migrations). It never touches the filesystem, which lets the
// hot-reload path use it without LoadConfig's create-if-missing side
// effect.
func parseConfig(data []byte) (Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Colleague work/sleep hours are not defaulted here: unset (nil)
	// fields fall back to defaults via the Get* accessors, so an explicit
	// 0 (midnight) in the config is preserved rather than rewritten.
	//
	// One exception: configs written by older versions always contained
	// explicit hours, with the pair (0, 0) as their "use defaults"
	// sentinel. A 0-0 range is empty and cannot be meant literally, so
	// map it back to unset. Genuine midnight bounds (e.g. sleep 22-0)
	// contain a non-zero value and are untouched.
	for i := range config.Colleagues {
		c := &config.Colleagues[i]
		if c.SleepStart != nil && c.SleepEnd != nil && *c.SleepStart == 0 && *c.SleepEnd == 0 {
			c.SleepStart, c.SleepEnd = nil, nil
		}
		if c.WorkStart != nil && c.WorkEnd != nil && *c.WorkStart == 0 && *c.WorkEnd == 0 {
			c.WorkStart, c.WorkEnd = nil, nil
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

	// Set default color scheme if not specified
	if config.ColorScheme == "" {
		config.ColorScheme = "classic"
	}

	// Set default timeline mode if not specified
	if config.TimelineMode == "" {
		config.TimelineMode = "individual"
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
