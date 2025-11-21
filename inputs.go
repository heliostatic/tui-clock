package main

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// newNameInput creates a fresh, properly configured name input
func newNameInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Colleague name"
	input.CharLimit = 50
	input.Width = 50
	input.Prompt = ""
	return input
}

// newNameInputWithValue creates a name input pre-filled with a value
func newNameInputWithValue(value string) textinput.Model {
	input := newNameInput()
	input.SetValue(value)
	return input
}

// exitToNormal returns the model to normal mode and clears state
func (m *Model) exitToNormal() {
	m.inputMode = ModeNormal
	m.nameInput.Blur()
	m.errorMsg = ""
}

// enterSearchMode prepares the model for timezone search (add flow)
func (m *Model) enterSearchMode() {
	m.inputMode = ModeSearchTimezone
	m.nameInput.Blur()
	m.searchQuery = ""
	m.updateSearchResults()
}

// enterEditSearchMode prepares the model for timezone search (edit flow)
func (m *Model) enterEditSearchMode() {
	m.inputMode = ModeEditSearchTimezone
	m.nameInput.Blur()
	m.searchQuery = ""
	m.updateSearchResults()
}

// handleSearchNavigation handles up/down/typing in search mode
// Returns true if the key was handled, false otherwise
func (m *Model) handleSearchNavigation(key string) bool {
	const maxVisible = 10

	switch key {
	case "up", "k":
		if m.searchCursor > 0 {
			m.searchCursor--
			if m.searchCursor < m.searchScrollOffset {
				m.searchScrollOffset = m.searchCursor
			}
		}
		return true

	case "down", "j":
		if m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
			if m.searchCursor >= m.searchScrollOffset+maxVisible {
				m.searchScrollOffset = m.searchCursor - maxVisible + 1
			}
		}
		return true

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateSearchResults()
		}
		return true

	default:
		// Handle printable ASCII character input
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.searchQuery += key
			m.updateSearchResults()
			return true
		}
		return false
	}
}
