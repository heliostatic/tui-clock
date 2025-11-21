package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

// Colleague represents a person with their timezone information
type Colleague struct {
	Name      string `yaml:"name"`
	Timezone  string `yaml:"timezone"`
	WorkStart int    `yaml:"work_start"` // Hour in 24h format (e.g., 9 for 9am)
	WorkEnd   int    `yaml:"work_end"`   // Hour in 24h format (e.g., 17 for 5pm)
}

// Config represents the application configuration
type Config struct {
	TimeFormat            string      `yaml:"time_format"`             // "12h" or "24h"
	LocationDisplayFormat string      `yaml:"location_display_format"` // "auto", "city", "timezone", "abbreviation"
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
)

// Application constants
const (
	AutoHideTimeout  = 3 * time.Second // Time before selection indicator hides
	DefaultWorkStart = 9               // Default work start hour (9am)
	DefaultWorkEnd   = 17              // Default work end hour (5pm)
	MaxVisible       = 8               // Maximum colleagues visible at once
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
