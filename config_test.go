package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.TimeFormat != "24h" {
		t.Errorf("Expected default time format '24h', got '%s'", config.TimeFormat)
	}

	if config.LocationDisplayFormat != "auto" {
		t.Errorf("Expected default location display format 'auto', got '%s'", config.LocationDisplayFormat)
	}

	if len(config.Colleagues) == 0 {
		t.Error("Expected default config to have example colleagues")
	}

	// Verify default colleagues have working hours set
	for i, colleague := range config.Colleagues {
		if colleague.WorkStart == 0 || colleague.WorkEnd == 0 {
			t.Errorf("Colleague %d missing work hours: start=%d, end=%d",
				i, colleague.WorkStart, colleague.WorkEnd)
		}
		if colleague.WorkStart >= colleague.WorkEnd {
			t.Errorf("Colleague %d has invalid work hours: start=%d >= end=%d",
				i, colleague.WorkStart, colleague.WorkEnd)
		}
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	// Create a test config
	testConfig := Config{
		TimeFormat:            "12h",
		LocationDisplayFormat: "city",
		Colleagues: []Colleague{
			{
				Name:      "Test User (Tokyo)",
				Timezone:  "Asia/Tokyo",
				WorkStart: 10,
				WorkEnd:   18,
			},
		},
	}

	// Save config
	err := SaveConfig(configPath, testConfig)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config back
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded config matches
	if loadedConfig.TimeFormat != testConfig.TimeFormat {
		t.Errorf("TimeFormat mismatch: got '%s', want '%s'",
			loadedConfig.TimeFormat, testConfig.TimeFormat)
	}

	if loadedConfig.LocationDisplayFormat != testConfig.LocationDisplayFormat {
		t.Errorf("LocationDisplayFormat mismatch: got '%s', want '%s'",
			loadedConfig.LocationDisplayFormat, testConfig.LocationDisplayFormat)
	}

	if len(loadedConfig.Colleagues) != len(testConfig.Colleagues) {
		t.Errorf("Colleagues count mismatch: got %d, want %d",
			len(loadedConfig.Colleagues), len(testConfig.Colleagues))
	}

	if len(loadedConfig.Colleagues) > 0 {
		colleague := loadedConfig.Colleagues[0]
		if colleague.Name != "Test User (Tokyo)" {
			t.Errorf("Colleague name mismatch: got '%s', want 'Test User (Tokyo)'", colleague.Name)
		}
		if colleague.Timezone != "Asia/Tokyo" {
			t.Errorf("Colleague timezone mismatch: got '%s', want 'Asia/Tokyo'", colleague.Timezone)
		}
		if colleague.WorkStart != 10 || colleague.WorkEnd != 18 {
			t.Errorf("Colleague work hours mismatch: got %d-%d, want 10-18",
				colleague.WorkStart, colleague.WorkEnd)
		}
	}
}

func TestLoadConfigNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Loading nonexistent config should create default config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should create default config, got error: %v", err)
	}

	// Verify it's a default config
	if config.TimeFormat != "24h" {
		t.Errorf("Expected default time format, got '%s'", config.TimeFormat)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Default config file was not created")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial_config.yaml")

	// Create a config with missing optional fields
	partialYAML := `colleagues:
  - name: "Test"
    timezone: "UTC"
`
	err := os.WriteFile(configPath, []byte(partialYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check defaults are applied
	if config.TimeFormat != "24h" {
		t.Errorf("Expected default time format '24h', got '%s'", config.TimeFormat)
	}

	if config.LocationDisplayFormat != "auto" {
		t.Errorf("Expected default location display format 'auto', got '%s'", config.LocationDisplayFormat)
	}

	if len(config.Colleagues) > 0 {
		colleague := config.Colleagues[0]
		if colleague.WorkStart != 9 {
			t.Errorf("Expected default work start 9, got %d", colleague.WorkStart)
		}
		if colleague.WorkEnd != 17 {
			t.Errorf("Expected default work end 17, got %d", colleague.WorkEnd)
		}
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Loading invalid config should return error
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid config, got nil")
	}
}
