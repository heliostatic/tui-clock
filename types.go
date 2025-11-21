package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

// Colleague represents a person with their timezone information
type Colleague struct {
	Name       string `yaml:"name"`
	Timezone   string `yaml:"timezone"`
	WorkStart  int    `yaml:"work_start"`  // Hour in 24h format (e.g., 9 for 9am)
	WorkEnd    int    `yaml:"work_end"`    // Hour in 24h format (e.g., 17 for 5pm)
	SleepStart int    `yaml:"sleep_start"` // Hour in 24h format (e.g., 23 for 11pm), 0 = use default
	SleepEnd   int    `yaml:"sleep_end"`   // Hour in 24h format (e.g., 7 for 7am), 0 = use default
}

// GetWorkStart returns the work start hour, using the default if not set
func (c Colleague) GetWorkStart() int {
	if c.WorkStart == 0 {
		return DefaultWorkStart
	}
	return c.WorkStart
}

// GetWorkEnd returns the work end hour, using the default if not set
func (c Colleague) GetWorkEnd() int {
	if c.WorkEnd == 0 {
		return DefaultWorkEnd
	}
	return c.WorkEnd
}

// GetSleepStart returns the sleep start hour, using the default if not set
func (c Colleague) GetSleepStart() int {
	if c.SleepStart == 0 {
		return DefaultSleepStart
	}
	return c.SleepStart
}

// GetSleepEnd returns the sleep end hour, using the default if not set
func (c Colleague) GetSleepEnd() int {
	if c.SleepEnd == 0 {
		return DefaultSleepEnd
	}
	return c.SleepEnd
}

// Config represents the application configuration
type Config struct {
	TimeFormat            string      `yaml:"time_format"`             // "12h" or "24h"
	LocationDisplayFormat string      `yaml:"location_display_format"` // "auto", "city", "timezone", "abbreviation"
	ColorScheme           string      `yaml:"color_scheme"`            // "classic", "dark", "high-contrast"
	TimelineMode          string      `yaml:"timeline_mode"`           // "individual", "shared"
	Colleagues            []Colleague `yaml:"colleagues"`
}

// ColleagueTime holds computed time information for display
type ColleagueTime struct {
	Colleague     Colleague
	CurrentTime   time.Time
	Offset        string // e.g., "+5h", "-8h", "same"
	IsWorkingTime bool
	IsWeekend     bool
}

// InputMode represents the current input state
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeAddName
	ModeSearchTimezone // New mode for searching/selecting timezone
	ModeEditName
	ModeEditSearchTimezone // Edit mode for timezone search
	ModeHelp
	ModeTimeline // Timeline visualization mode
)

// Application constants
const (
	AutoHideTimeout   = 3 * time.Second // Time before selection indicator hides
	DefaultWorkStart  = 9               // Default work start hour (9am)
	DefaultWorkEnd    = 17              // Default work end hour (5pm)
	DefaultSleepStart = 23              // Default sleep start hour (11pm)
	DefaultSleepEnd   = 7               // Default sleep end hour (7am)
	MaxVisible        = 8               // Maximum colleagues visible at once

	// Timeline visualization constants
	MinBarWidth    = 24 // Minimum bar width (1 char per hour)
	IdealBarWidth  = 48 // Ideal bar width (2 chars per hour)
	MaxBarWidth    = 72 // Maximum bar width (3 chars per hour)
	NameFieldWidth = 25 // Width for colleague name field
	TimeFieldWidth = 12 // Width for time field
)

// Model represents the Bubbletea application state
type Model struct {
	config          Config
	configPath      string
	colleagues      []ColleagueTime
	localTimezone   *time.Location
	cursor          int       // Selected item index (or last known position)
	selectionActive bool      // Whether selection is visually shown
	lastActionTime  time.Time // Time of last user action (for auto-hide)
	scrollOffset    int       // Scroll position
	inputMode       InputMode // Current input mode
	nameInput       textinput.Model
	editIndex       int    // Index of colleague being edited
	errorMsg        string // Error message to display
	width           int    // Terminal width
	height          int    // Terminal height

	// Timezone search state
	searchQuery        string         // Current search query (what user typed)
	searchResults      []SearchResult // Filtered search results
	searchCursor       int            // Selected result index
	searchScrollOffset int            // Scroll position in search results
}
