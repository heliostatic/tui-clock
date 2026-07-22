package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg builds a tea.KeyMsg for tests: named keys map to their key
// types, anything else becomes a KeyRunes message (possibly multi-rune,
// as bubbletea delivers for pasted or fast-typed input)
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestNewNameInput(t *testing.T) {
	input := newNameInput()

	if input.Placeholder != "Colleague name" {
		t.Errorf("Expected placeholder 'Colleague name', got '%s'", input.Placeholder)
	}

	if input.CharLimit != 50 {
		t.Errorf("Expected char limit 50, got %d", input.CharLimit)
	}

	if input.Width != 50 {
		t.Errorf("Expected width 50, got %d", input.Width)
	}

	if input.Prompt != "" {
		t.Errorf("Expected empty prompt, got '%s'", input.Prompt)
	}
}

func TestNewNameInputWithValue(t *testing.T) {
	testValue := "Test User"
	input := newNameInputWithValue(testValue)

	if input.Value() != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, input.Value())
	}

	// Should still have all the standard properties
	if input.Placeholder != "Colleague name" {
		t.Errorf("Expected placeholder 'Colleague name', got '%s'", input.Placeholder)
	}

	if input.CharLimit != 50 {
		t.Errorf("Expected char limit 50, got %d", input.CharLimit)
	}
}

func TestExitToNormal(t *testing.T) {
	m := Model{
		inputMode: ModeAddName,
		errorMsg:  "Some error",
	}
	m.nameInput = newNameInput()
	m.nameInput.Focus()

	m.exitToNormal()

	if m.inputMode != ModeNormal {
		t.Errorf("Expected mode ModeNormal, got %v", m.inputMode)
	}

	if m.errorMsg != "" {
		t.Errorf("Expected error message to be cleared, got '%s'", m.errorMsg)
	}

	// Note: Blur() doesn't expose a way to check focus state externally,
	// so we can't directly test that, but we called it
}

func TestEnterSearchMode(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config, "/tmp/test.yaml")
	m.searchQuery = "old query"
	m.searchResults = []SearchResult{{}} // Non-empty

	m.enterSearchMode()

	if m.inputMode != ModeSearchTimezone {
		t.Errorf("Expected mode ModeSearchTimezone, got %v", m.inputMode)
	}

	if m.searchQuery != "" {
		t.Errorf("Expected search query to be cleared, got '%s'", m.searchQuery)
	}

	// updateSearchResults() should have been called
	// With empty query, it returns all cities
	if len(m.searchResults) == 0 {
		t.Error("Expected search results to be populated with all cities")
	}
}

func TestEnterEditSearchMode(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config, "/tmp/test.yaml")
	m.searchQuery = "old query"

	m.enterEditSearchMode()

	if m.inputMode != ModeEditSearchTimezone {
		t.Errorf("Expected mode ModeEditSearchTimezone, got %v", m.inputMode)
	}

	if m.searchQuery != "" {
		t.Errorf("Expected search query to be cleared, got '%s'", m.searchQuery)
	}
}

func TestHandleSearchNavigation(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config, "/tmp/test.yaml")

	// Populate with some search results
	m.searchResults = SearchTimezones("") // Get all results
	if len(m.searchResults) < 5 {
		t.Fatal("Need at least 5 search results for this test")
	}
	m.searchCursor = 0
	m.searchScrollOffset = 0

	tests := []struct {
		name           string
		key            string
		expectHandled  bool
		checkCursor    bool
		expectedCursor int
	}{
		{"Down arrow moves cursor", "down", true, true, 1},
		{"Up arrow moves cursor", "up", true, true, 0},
		{"j types into query, cursor unchanged", "j", true, true, 0},
		{"k types into query, cursor unchanged", "k", true, true, 0},
		{"Backspace handled", "backspace", true, false, -1},
		{"Printable char handled", "a", true, false, -1},
		{"Non-printable ignored", "ctrl+c", false, false, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			m.searchCursor = 0
			m.searchQuery = ""

			handled := m.handleSearchNavigation(keyMsg(tt.key))

			if handled != tt.expectHandled {
				t.Errorf("handleSearchNavigation(%s) returned %v, expected %v",
					tt.key, handled, tt.expectHandled)
			}

			if tt.checkCursor && m.searchCursor != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, m.searchCursor)
			}
		})
	}
}

func TestHandleSearchNavigationBounds(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config, "/tmp/test.yaml")
	m.searchResults = SearchTimezones("")

	if len(m.searchResults) == 0 {
		t.Fatal("Need search results for this test")
	}

	// Test up at top boundary
	m.searchCursor = 0
	m.handleSearchNavigation(keyMsg("up"))
	if m.searchCursor != 0 {
		t.Errorf("Cursor should stay at 0 when going up from top, got %d", m.searchCursor)
	}

	// Test down at bottom boundary
	m.searchCursor = len(m.searchResults) - 1
	m.handleSearchNavigation(keyMsg("down"))
	if m.searchCursor != len(m.searchResults)-1 {
		t.Errorf("Cursor should stay at max when going down from bottom, got %d", m.searchCursor)
	}
}

func TestHandleSearchNavigationTyping(t *testing.T) {
	config := DefaultConfig()
	m := NewModel(config, "/tmp/test.yaml")
	m.searchQuery = ""

	// Type some characters
	m.handleSearchNavigation(keyMsg("l"))
	m.handleSearchNavigation(keyMsg("o"))
	m.handleSearchNavigation(keyMsg("n"))

	if m.searchQuery != "lon" {
		t.Errorf("Expected search query 'lon', got '%s'", m.searchQuery)
	}

	// Backspace should remove last character
	m.handleSearchNavigation(keyMsg("backspace"))
	if m.searchQuery != "lo" {
		t.Errorf("Expected search query 'lo' after backspace, got '%s'", m.searchQuery)
	}

	// Letters used elsewhere for vim-style navigation must still be typeable
	m.searchQuery = ""
	for _, key := range []string{"t", "o", "k", "y", "o"} {
		m.handleSearchNavigation(keyMsg(key))
	}
	if m.searchQuery != "tokyo" {
		t.Errorf("Expected search query 'tokyo', got '%s'", m.searchQuery)
	}

	// Pasted or fast-typed input arrives as one multi-rune KeyRunes
	// message; it must not be dropped
	m.searchQuery = ""
	handled := m.handleSearchNavigation(keyMsg("tokyo"))
	if !handled {
		t.Error("Expected multi-rune input to be handled")
	}
	if m.searchQuery != "tokyo" {
		t.Errorf("Expected search query 'tokyo' after paste, got '%s'", m.searchQuery)
	}

	// Space arrives as KeySpace, not KeyRunes, and must still be typeable
	m.searchQuery = "new"
	m.handleSearchNavigation(keyMsg(" "))
	if m.searchQuery != "new " {
		t.Errorf("Expected search query 'new ' after space, got '%s'", m.searchQuery)
	}
}
