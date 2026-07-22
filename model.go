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

// tick returns a command that sends a TickMsg at the next wall-clock
// second boundary; ticking a fixed interval after processing would
// slowly drift and skip displayed seconds
func tick() tea.Cmd {
	untilNextSecond := time.Until(time.Now().Truncate(time.Second).Add(time.Second))
	return tea.Tick(untilNextSecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// newColleague creates a new colleague; unset hour fields use the
// defaults via the Get* accessors
func newColleague(name, timezone string) Colleague {
	return Colleague{
		Name:     name,
		Timezone: timezone,
	}
}

// displayNow returns the local time the timeline should render: the
// real current time plus any scrub offset
func (m Model) displayNow() time.Time {
	return time.Now().In(m.localTimezone).Add(m.timeOffset)
}

// scrubbed returns a copy of ct shifted by the scrub offset, with the
// time-dependent flags recomputed for the shifted moment (scrubbing
// can cross midnight and change the weekday)
func (m Model) scrubbed(ct ColleagueTime) ColleagueTime {
	if m.timeOffset == 0 || ct.InvalidTimezone {
		return ct
	}
	ct.CurrentTime = ct.CurrentTime.Add(m.timeOffset)
	ct.IsWeekend = ct.CurrentTime.Weekday() == time.Saturday || ct.CurrentTime.Weekday() == time.Sunday
	ct.IsWorkingTime = !ct.IsWeekend &&
		isInTimeRange(ct.CurrentTime.Hour(), ct.Colleague.GetWorkStart(), ct.Colleague.GetWorkEnd())
	return ct
}

// updateColleagueTimes recomputes all colleague times
func (m *Model) updateColleagueTimes() {
	m.colleagues = ComputeColleagueTimes(m.config.Colleagues, m.localTimezone)
}

// saveConfig saves the current config to file
func (m *Model) saveConfig() error {
	return SaveConfig(m.configPath, m.config)
}

// deleteColleague removes a colleague and saves config
func (m *Model) deleteColleague(index int) error {
	if index < 0 || index >= len(m.config.Colleagues) {
		return nil
	}

	m.config.Colleagues = append(m.config.Colleagues[:index], m.config.Colleagues[index+1:]...)
	m.updateColleagueTimes()

	// Adjust cursor if necessary (cursor indexes the display list)
	if m.cursor >= len(m.colleagues) && m.cursor > 0 {
		m.cursor--
	}

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

	colleague := newColleague(finalName, result.City.Timezone)

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
