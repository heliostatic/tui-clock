package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeConfigWithMtime writes content to path and bumps its mtime to a
// distinct future value, so tests don't depend on filesystem timestamp
// granularity.
func writeConfigWithMtime(t *testing.T, path, content string, mtime time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("Failed to set mtime: %v", err)
	}
}

func newReloadTestModel(t *testing.T) (Model, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	config, err := LoadConfig(path) // Creates the default config file
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	return NewModel(config, path), path
}

func TestMaybeReloadConfigPicksUpExternalEdit(t *testing.T) {
	m, path := newReloadTestModel(t)

	external := `time_format: "12h"
colleagues:
  - name: "External Edit"
    timezone: "Asia/Tokyo"
`
	writeConfigWithMtime(t, path, external, time.Now().Add(time.Hour))

	m.maybeReloadConfig()

	if m.config.TimeFormat != "12h" {
		t.Errorf("TimeFormat = %q, want 12h from external edit", m.config.TimeFormat)
	}
	if len(m.colleagues) != 1 || m.colleagues[0].Colleague.Name != "External Edit" {
		t.Errorf("Colleague list not reloaded: %+v", m.colleagues)
	}
}

func TestMaybeReloadConfigIgnoresOwnSave(t *testing.T) {
	m, _ := newReloadTestModel(t)

	m.config.TimeFormat = "12h"
	if err := m.saveConfig(); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	before := m.configMtime
	m.maybeReloadConfig()
	if !m.configMtime.Equal(before) {
		t.Error("maybeReloadConfig treated our own save as an external edit")
	}
	if m.config.TimeFormat != "12h" {
		t.Errorf("In-memory change lost: TimeFormat = %q", m.config.TimeFormat)
	}
}

func TestMaybeReloadConfigDeferredDuringEdit(t *testing.T) {
	m, path := newReloadTestModel(t)

	external := `colleagues:
  - name: "Should Wait"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, external, time.Now().Add(time.Hour))

	// A modal flow holds editIndex into the config: reload must wait
	m.inputMode = ModeEditWorkHours
	m.maybeReloadConfig()
	if len(m.config.Colleagues) == 1 {
		t.Fatal("Reload must be deferred while an edit flow is open")
	}

	// Back in normal mode the reload goes through
	m.inputMode = ModeNormal
	m.maybeReloadConfig()
	if len(m.config.Colleagues) != 1 || m.config.Colleagues[0].Name != "Should Wait" {
		t.Errorf("Expected deferred reload to apply in normal mode, got %+v", m.config.Colleagues)
	}
}

func TestMaybeReloadConfigSurvivesTornWrite(t *testing.T) {
	m, path := newReloadTestModel(t)
	originalCount := len(m.config.Colleagues)

	// A partial editor write: invalid YAML
	writeConfigWithMtime(t, path, "colleagues: [inva", time.Now().Add(time.Hour))
	m.maybeReloadConfig()
	if len(m.config.Colleagues) != originalCount {
		t.Fatal("Torn write must not clobber the running config")
	}

	// The write completes on a later tick: now it reloads
	valid := `colleagues:
  - name: "Recovered"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, valid, time.Now().Add(2*time.Hour))
	m.maybeReloadConfig()
	if len(m.config.Colleagues) != 1 || m.config.Colleagues[0].Name != "Recovered" {
		t.Errorf("Expected reload after the file became valid, got %+v", m.config.Colleagues)
	}

	// The torn write must not have been overwritten by a save of the old
	// config: the file should still contain what the editor wrote
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != valid {
		t.Error("Reload path must never write to the config file")
	}
}

func TestMaybeReloadConfigClampsSelection(t *testing.T) {
	m, path := newReloadTestModel(t)
	m.cursor = len(m.colleagues) - 1 // Select the last of the 3 defaults
	m.selectionActive = true

	external := `colleagues:
  - name: "Only One"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, external, time.Now().Add(time.Hour))
	m.maybeReloadConfig()

	if m.cursor != 0 {
		t.Errorf("Cursor = %d, want clamped to 0", m.cursor)
	}

	// Shrink to empty: cursor becomes no-selection
	writeConfigWithMtime(t, path, "colleagues: []\n", time.Now().Add(2*time.Hour))
	m.maybeReloadConfig()
	if m.cursor != -1 || m.selectionActive {
		t.Errorf("Empty list: cursor = %d selectionActive = %v, want -1/false", m.cursor, m.selectionActive)
	}
}
