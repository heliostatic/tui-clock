package main

import (
	"os"
	"path/filepath"
	"strings"
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

func TestMaybeReloadConfigResetsSelection(t *testing.T) {
	m, path := newReloadTestModel(t)
	m.cursor = len(m.colleagues) - 1 // Select the last of the 3 defaults
	m.selectionActive = true
	m.scrollOffset = 2

	// The external edit may reorder entries, so keeping the cursor
	// would silently select a different colleague; the selection is
	// dropped instead (matching in-app delete behavior)
	external := `colleagues:
  - name: "Only One"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, external, time.Now().Add(time.Hour))
	m.maybeReloadConfig()

	if m.cursor != -1 || m.selectionActive {
		t.Errorf("Cursor = %d selectionActive = %v, want -1/false after reload", m.cursor, m.selectionActive)
	}
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want clamped to 0", m.scrollOffset)
	}
}

func TestMaybeReloadConfigTriggersOnOlderMtime(t *testing.T) {
	m, path := newReloadTestModel(t)

	// A timestamp-preserving restore (cp -p, rsync -t, tar -x) can give
	// the file an mtime EARLIER than the one we recorded; any mtime
	// difference must trigger a reload, not just newer ones. The edit
	// is deliberately the same byte length as the current file so the
	// size fallback can't mask a wrong mtime comparison (e.g. After
	// instead of Equal).
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	sameSize := strings.Replace(string(data), "Alice", "Elise", 1)
	if len(sameSize) != len(data) {
		t.Fatal("Test setup: replacement must preserve file size")
	}
	writeConfigWithMtime(t, path, sameSize, m.configMtime.Add(-time.Hour))
	m.maybeReloadConfig()

	found := false
	for _, c := range m.config.Colleagues {
		if strings.Contains(c.Name, "Elise") {
			found = true
		}
	}
	if !found {
		t.Errorf("Same-size older-mtime edit not reloaded: %+v", m.config.Colleagues)
	}
}

func TestMaybeReloadConfigDetectsSameMtimeRewrite(t *testing.T) {
	m, path := newReloadTestModel(t)

	// On coarse-timestamp filesystems an external write can land in the
	// same timestamp quantum as our own save; the size comparison must
	// still catch it (unless sizes also collide, which we accept)
	external := `colleagues:
  - name: "Same Quantum"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, external, m.configMtime)
	m.maybeReloadConfig()

	if len(m.config.Colleagues) != 1 || m.config.Colleagues[0].Name != "Same Quantum" {
		t.Errorf("Same-mtime different-size edit not reloaded: %+v", m.config.Colleagues)
	}
}

func TestMaybeReloadConfigNeverResurrectsDeletedFile(t *testing.T) {
	m, path := newReloadTestModel(t)
	originalCount := len(m.config.Colleagues)

	// Editors with rename-style saves briefly remove the file; the
	// reload path must neither adopt defaults nor recreate the file
	if err := os.Remove(path); err != nil {
		t.Fatalf("Failed to remove config: %v", err)
	}
	m.maybeReloadConfig()

	if len(m.config.Colleagues) != originalCount {
		t.Error("Reload must keep the running config when the file vanishes")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Reload must not recreate a deleted config file")
	}

	// When the rename completes, the new content is picked up
	external := `colleagues:
  - name: "Renamed Into Place"
    timezone: "UTC"
`
	writeConfigWithMtime(t, path, external, time.Now().Add(time.Hour))
	m.maybeReloadConfig()
	if len(m.config.Colleagues) != 1 || m.config.Colleagues[0].Name != "Renamed Into Place" {
		t.Errorf("Expected reload after rename completed, got %+v", m.config.Colleagues)
	}
}
