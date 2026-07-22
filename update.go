package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		m.maybeReloadConfig()
		m.updateColleagueTimes()

		// Auto-hide selection after inactivity timeout
		if m.selectionActive && time.Since(m.lastActionTime) > AutoHideTimeout {
			m.selectionActive = false
		}

		return m, tick()

	default:
		return m, nil
	}
}

// handleKeyPress handles keyboard input based on current mode
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit key
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.inputMode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeAddName:
		return m.handleAddNameMode(msg)
	case ModeSearchTimezone:
		return m.handleSearchTimezoneMode(msg)
	case ModeEditName:
		return m.handleEditNameMode(msg)
	case ModeEditSearchTimezone:
		return m.handleEditSearchTimezoneMode(msg)
	case ModeEditWorkHours:
		return m.handleEditWorkHoursMode(msg)
	case ModeEditSleepHours:
		return m.handleEditSleepHoursMode(msg)
	case ModeHelp:
		return m.handleHelpMode(msg)
	case ModeTimeline:
		return m.handleTimelineMode(msg)
	default:
		return m, nil
	}
}

// handleNormalMode handles keys in normal browsing mode
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit

	case "up", "k":
		// If selection is hidden (inactive), reactivate it first without moving
		if m.reactivateSelection() {
			return m, nil
		}

		// If no selection at all, activate at first item
		if m.cursor == -1 && len(m.colleagues) > 0 {
			m.cursor = 0
			m.activateSelection()
		} else if m.cursor > 0 {
			m.cursor--
			m.lastActionTime = time.Now()
			// Adjust scroll if cursor goes above visible area
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}

	case "down", "j":
		// If selection is hidden (inactive), reactivate it first without moving
		if m.reactivateSelection() {
			return m, nil
		}

		// If no selection at all, activate at first item
		if m.cursor == -1 && len(m.colleagues) > 0 {
			m.cursor = 0
			m.activateSelection()
		} else if m.cursor < len(m.colleagues)-1 {
			m.cursor++
			m.lastActionTime = time.Now()
			// Adjust scroll if cursor goes below visible area
			if m.cursor >= m.scrollOffset+MaxVisible {
				m.scrollOffset = m.cursor - MaxVisible + 1
			}
		}

	case "a":
		// Add new colleague
		m.inputMode = ModeAddName
		m.nameInput = newNameInput()
		m.nameInput.Focus()
		m.errorMsg = ""

	case "d":
		// If selection is hidden (inactive), reactivate it first without deleting
		if m.reactivateSelection() {
			return m, nil
		}

		// Delete selected colleague (only if something is selected and active).
		// Use the config index, not the display cursor: the display list skips
		// entries with invalid timezones, so the two can diverge.
		if m.cursor >= 0 && m.cursor < len(m.colleagues) && m.selectionActive {
			if err := m.deleteColleague(m.colleagues[m.cursor].ConfigIndex); err != nil {
				m.errorMsg = err.Error()
			} else {
				// Return to no selection after delete
				m.cursor = -1
				m.selectionActive = false
			}
		}

	case "e":
		// If selection is hidden (inactive), reactivate it first without editing
		if m.reactivateSelection() {
			return m, nil
		}

		// Edit selected colleague (only if something is selected and active).
		// editIndex is a config index; see the delete handler for why.
		if m.cursor >= 0 && m.cursor < len(m.colleagues) && m.selectionActive {
			m.inputMode = ModeEditName
			m.editIndex = m.colleagues[m.cursor].ConfigIndex
			m.nameInput = newNameInputWithValue(m.colleagues[m.cursor].Colleague.Name)
			m.nameInput.Focus()
			m.errorMsg = ""
		}

	case "w":
		// If selection is hidden (inactive), reactivate it first without editing
		if m.reactivateSelection() {
			return m, nil
		}

		// Edit selected colleague's work/sleep hours
		if m.cursor >= 0 && m.cursor < len(m.colleagues) && m.selectionActive {
			ct := m.colleagues[m.cursor]
			m.inputMode = ModeEditWorkHours
			m.editIndex = ct.ConfigIndex
			m.pendingWorkAction = hourRangeKeep
			m.nameInput = newHourRangeInput(fmt.Sprintf("%d-%d",
				ct.Colleague.GetWorkStart(), ct.Colleague.GetWorkEnd()))
			m.nameInput.Focus()
			m.errorMsg = ""
		}

	case "f":
		// Toggle time format
		if err := m.toggleTimeFormat(); err != nil {
			m.errorMsg = err.Error()
		}

	case "?", "h":
		// Show help
		m.inputMode = ModeHelp

	case "t":
		// Enter timeline mode
		m.inputMode = ModeTimeline
	}

	return m, nil
}

// handleAddNameMode handles input when adding a new colleague's name
func (m Model) handleAddNameMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.nameInput.Value() != "" {
			m.enterSearchMode()
		}
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// handleSearchTimezoneMode handles timezone search when adding a colleague
func (m Model) handleSearchTimezoneMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Select the currently highlighted timezone
		if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
			result := m.searchResults[m.searchCursor]
			if err := m.addColleagueFromSearch(m.nameInput.Value(), result); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.exitToNormal()
				m.cursor = len(m.colleagues) - 1
				// Scroll the new entry into view (it's appended last)
				if m.cursor >= m.scrollOffset+MaxVisible {
					m.scrollOffset = m.cursor - MaxVisible + 1
				}
				m.activateSelection()
			}
		}
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil

	default:
		// Handle search navigation (up/down/typing)
		m.handleSearchNavigation(msg)
		return m, nil
	}
}

// handleEditNameMode handles input when editing a colleague's name
func (m Model) handleEditNameMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.nameInput.Value() != "" {
			m.enterEditSearchMode()
		}
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// handleEditSearchTimezoneMode handles timezone search when editing a colleague
func (m Model) handleEditSearchTimezoneMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Select the currently highlighted timezone
		if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
			result := m.searchResults[m.searchCursor]
			if err := m.editColleagueFromSearch(m.editIndex, m.nameInput.Value(), result); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.exitToNormal()
				m.activateSelection()
			}
		}
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil

	default:
		// Handle search navigation (up/down/typing)
		m.handleSearchNavigation(msg)
		return m, nil
	}
}

// handleEditWorkHoursMode handles the work-hours step of hour editing.
// The parsed result is only staged; nothing touches the config until
// the sleep step is confirmed, so Esc at any point truly cancels.
func (m Model) handleEditWorkHoursMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		action, start, end, err := parseHourRange(m.nameInput.Value())
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		m.pendingWorkAction = action
		m.pendingWorkStart = start
		m.pendingWorkEnd = end

		// Continue to the sleep-hours step
		if m.editIndex >= 0 && m.editIndex < len(m.config.Colleagues) {
			c := m.config.Colleagues[m.editIndex]
			m.inputMode = ModeEditSleepHours
			m.nameInput = newHourRangeInput(fmt.Sprintf("%d-%d", c.GetSleepStart(), c.GetSleepEnd()))
			m.nameInput.Focus()
			m.errorMsg = ""
		} else {
			m.exitToNormal()
		}
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// handleEditSleepHoursMode handles the sleep-hours step of hour
// editing; confirming it applies the staged work hours and the sleep
// hours together
func (m Model) handleEditSleepHoursMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		action, start, end, err := parseHourRange(m.nameInput.Value())
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		switch m.pendingWorkAction {
		case hourRangeReset:
			err = m.applyWorkHours(m.editIndex, nil, nil)
		case hourRangeSet:
			err = m.applyWorkHours(m.editIndex, HourPtr(m.pendingWorkStart), HourPtr(m.pendingWorkEnd))
		}
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		switch action {
		case hourRangeReset:
			err = m.applySleepHours(m.editIndex, nil, nil)
		case hourRangeSet:
			err = m.applySleepHours(m.editIndex, HourPtr(start), HourPtr(end))
		}
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		m.exitToNormal()
		m.activateSelection()
		return m, nil

	case "esc":
		m.exitToNormal()
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// handleHelpMode handles input in help screen
func (m Model) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.inputMode = ModeNormal
	return m, nil
}

// handleTimelineMode handles input in timeline visualization mode
func (m Model) handleTimelineMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "t":
		// Return to normal mode (scrub does not persist across modes)
		m.inputMode = ModeNormal
		m.timeOffset = 0

	case "esc":
		// First esc resets an active scrub; a second exits timeline mode
		if m.timeOffset != 0 {
			m.timeOffset = 0
		} else {
			m.inputMode = ModeNormal
		}

	case "left":
		// Scrub time backward one hour
		m.timeOffset -= time.Hour

	case "right":
		// Scrub time forward one hour
		m.timeOffset += time.Hour

	case "up", "k":
		// Scroll up
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}

	case "down", "j":
		// Scroll down
		maxScroll := max(len(m.colleagues)-MaxVisible, 0)
		if m.scrollOffset < maxScroll {
			m.scrollOffset++
		}

	case "c":
		// Cycle through color schemes (auto-discover from registered schemes)
		m.config.ColorScheme = GetNextColorScheme(m.config.ColorScheme)

		if err := m.saveConfig(); err != nil {
			m.errorMsg = err.Error()
		}

	case "m":
		// Toggle timeline mode
		if m.config.TimelineMode == "individual" {
			m.config.TimelineMode = "shared"
		} else {
			m.config.TimelineMode = "individual"
		}

		if err := m.saveConfig(); err != nil {
			m.errorMsg = err.Error()
		}

	case "?", "h":
		// Show help
		m.inputMode = ModeHelp
	}

	return m, nil
}
