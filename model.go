package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent every second to update the times
type TickMsg time.Time

// NewModel creates a new model with the given config
func NewModel(config Config, configPath string) Model {
	// Auto-detect local timezone
	localTz := time.Now().Location()

	// Create text input (will be replaced with fresh instance when entering add/edit mode)
	nameInput := newNameInput()

	m := Model{
		config:          config,
		configPath:      configPath,
		localTimezone:   localTz,
		inputMode:       ModeNormal,
		nameInput:       nameInput,
		cursor:          -1,    // Start with no selection for cleaner monitoring view
		selectionActive: false, // No visual selection initially
		lastActionTime:  time.Now(),
		scrollOffset:    0,
		// Initialize search state
		searchQuery:        "",
		searchResults:      []SearchResult{},
		searchCursor:       0,
		searchScrollOffset: 0,
	}

	// Compute initial times
	m.updateColleagueTimes()

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tick()
}

// tick returns a command that sends a TickMsg every second
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// updateColleagueTimes recomputes all colleague times
func (m *Model) updateColleagueTimes() {
	colleagues, _ := ComputeColleagueTimes(m.config.Colleagues, m.localTimezone, m.config.TimeFormat)
	m.colleagues = colleagues
}

// saveConfig saves the current config to file
func (m *Model) saveConfig() error {
	return SaveConfig(m.configPath, m.config)
}

// addColleague adds a new colleague and saves config
func (m *Model) addColleague(name, timezone string) error {
	// Validate timezone
	if err := ValidateTimezone(timezone); err != nil {
		return err
	}

	colleague := Colleague{
		Name:      name,
		Timezone:  timezone,
		WorkStart: 9,
		WorkEnd:   17,
	}

	m.config.Colleagues = append(m.config.Colleagues, colleague)
	m.updateColleagueTimes()
	return m.saveConfig()
}

// editColleague updates an existing colleague and saves config
func (m *Model) editColleague(index int, name, timezone string) error {
	if index < 0 || index >= len(m.config.Colleagues) {
		return nil
	}

	// Validate timezone
	if err := ValidateTimezone(timezone); err != nil {
		return err
	}

	m.config.Colleagues[index].Name = name
	m.config.Colleagues[index].Timezone = timezone
	m.updateColleagueTimes()
	return m.saveConfig()
}

// deleteColleague removes a colleague and saves config
func (m *Model) deleteColleague(index int) error {
	if index < 0 || index >= len(m.config.Colleagues) {
		return nil
	}

	m.config.Colleagues = append(m.config.Colleagues[:index], m.config.Colleagues[index+1:]...)

	// Adjust cursor if necessary
	if m.cursor >= len(m.config.Colleagues) && m.cursor > 0 {
		m.cursor--
	}

	m.updateColleagueTimes()
	return m.saveConfig()
}

// toggleTimeFormat switches between 12h and 24h format
func (m *Model) toggleTimeFormat() error {
	if m.config.TimeFormat == "12h" {
		m.config.TimeFormat = "24h"
	} else {
		m.config.TimeFormat = "12h"
	}
	m.updateColleagueTimes()
	return m.saveConfig()
}

// updateSearchResults updates the search results based on current query
func (m *Model) updateSearchResults() {
	m.searchResults = SearchTimezones(m.searchQuery)
	m.searchCursor = 0
	m.searchScrollOffset = 0
}

// addColleagueFromSearch adds a colleague using timezone search result
func (m *Model) addColleagueFromSearch(baseName string, result SearchResult) error {
	// Use smart append logic to format the name
	finalName := GetDisplayNameForColleague(baseName, result.City, m.searchQuery, m.config.LocationDisplayFormat)

	colleague := Colleague{
		Name:      finalName,
		Timezone:  result.City.Timezone,
		WorkStart: 9,
		WorkEnd:   17,
	}

	m.config.Colleagues = append(m.config.Colleagues, colleague)
	m.updateColleagueTimes()
	return m.saveConfig()
}

// editColleagueFromSearch updates a colleague using timezone search result
func (m *Model) editColleagueFromSearch(index int, baseName string, result SearchResult) error {
	if index < 0 || index >= len(m.config.Colleagues) {
		return nil
	}

	// Use smart append logic to format the name
	finalName := GetDisplayNameForColleague(baseName, result.City, m.searchQuery, m.config.LocationDisplayFormat)

	m.config.Colleagues[index].Name = finalName
	m.config.Colleagues[index].Timezone = result.City.Timezone
	m.updateColleagueTimes()
	return m.saveConfig()
}
