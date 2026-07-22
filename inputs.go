package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
// Navigation is arrow-keys only: letters like k/j must remain typeable
// since search is type-to-filter (e.g. "tokyo", "japan").
func (m *Model) handleSearchNavigation(msg tea.KeyMsg) bool {
	const maxVisible = 10

	switch msg.String() {
	case "up":
		if m.searchCursor > 0 {
			m.searchCursor--
			if m.searchCursor < m.searchScrollOffset {
				m.searchScrollOffset = m.searchCursor
			}
		}
		return true

	case "down":
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
		// Text input arrives as KeyRunes — possibly several runes in one
		// message when typed fast or pasted — or as KeySpace for a space.
		// Checking the message type (rather than string length) keeps
		// batched input from being silently dropped.
		if (msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace) && !msg.Alt {
			m.searchQuery += string(msg.Runes)
			m.updateSearchResults()
			return true
		}
		return false
	}
}

// reactivateSelection reactivates a hidden selection without performing any action
// Returns true if the keypress was consumed (selection was reactivated), false otherwise
func (m *Model) reactivateSelection() bool {
	if m.cursor >= 0 && !m.selectionActive {
		m.selectionActive = true
		m.lastActionTime = time.Now()
		return true // Consumed the keypress
	}
	return false // Continue processing
}

// activateSelection activates the selection and updates the last action time
func (m *Model) activateSelection() {
	m.selectionActive = true
	m.lastActionTime = time.Now()
}
