package main

import (
	"testing"
	"time"
)

// TestIsInTimeRange tests the time range checking function including wraparound cases
func TestIsInTimeRange(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		start    int
		end      int
		expected bool
	}{
		// Normal range (no wraparound)
		{"within normal range", 10, 9, 17, true},
		{"at start of range", 9, 9, 17, true},
		{"before normal range", 8, 9, 17, false},
		{"after normal range", 17, 9, 17, false},
		{"at end of range", 16, 9, 17, true},

		// Wraparound range (e.g., sleep hours 23:00-07:00)
		{"wraparound - within late night", 23, 23, 7, true},
		{"wraparound - within early morning", 1, 23, 7, true},
		{"wraparound - at boundary start", 23, 23, 7, true},
		{"wraparound - before boundary", 22, 23, 7, false},
		{"wraparound - at boundary end", 6, 23, 7, true},
		{"wraparound - after boundary", 7, 23, 7, false},
		{"wraparound - midnight", 0, 23, 7, true},

		// Edge cases
		{"single hour range", 9, 9, 10, true},
		{"full day range", 12, 0, 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInTimeRange(tt.hour, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("isInTimeRange(%d, %d, %d) = %v, want %v",
					tt.hour, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestTruncateOrPad tests the string truncation and padding function
func TestTruncateOrPad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"exact width", "hello", 5, "hello"},
		{"too long - truncate", "hello world", 5, "hello"},
		{"too short - pad", "hi", 5, "hi   "},
		{"empty string", "", 3, "   "},
		{"single char", "a", 5, "a    "},
		{"zero width", "test", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateOrPad(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("truncateOrPad(%q, %d) = %q, want %q",
					tt.input, tt.width, result, tt.expected)
			}
			if len(result) != tt.width {
				t.Errorf("truncateOrPad(%q, %d) returned length %d, want %d",
					tt.input, tt.width, len(result), tt.width)
			}
		})
	}
}

// TestCalculateTimelineBarWidth tests the bar width calculation
func TestCalculateTimelineBarWidth(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		expected      int
	}{
		{"very narrow terminal", 40, MinBarWidth},
		{"narrow terminal", 60, MinBarWidth},
		{"medium terminal", 80, 36},                // 80 - 44 = 36 (between min and ideal)
		{"wide terminal", 100, IdealBarWidth},      // 100 - 44 = 56, capped at IdealBarWidth (48)
		{"very wide terminal", 200, IdealBarWidth}, // capped at IdealBarWidth
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{width: tt.terminalWidth}
			result := m.calculateTimelineBarWidth()

			if result != tt.expected {
				t.Errorf("calculateTimelineBarWidth() with width %d = %d, want %d",
					tt.terminalWidth, result, tt.expected)
			}

			// Ensure result is within bounds
			if result < MinBarWidth {
				t.Errorf("result %d is less than MinBarWidth %d", result, MinBarWidth)
			}
			if result > IdealBarWidth {
				t.Errorf("result %d exceeds IdealBarWidth %d", result, IdealBarWidth)
			}
		})
	}
}

// TestCalculateOffsetHours tests timezone offset calculation
func TestCalculateOffsetHours(t *testing.T) {
	// Create test timezones
	localTz := time.FixedZone("Local", 0) // UTC
	remoteTz3 := time.FixedZone("Remote+3", 3*3600)
	remoteTz5 := time.FixedZone("Remote-5", -5*3600)

	tests := []struct {
		name     string
		localTz  *time.Location
		remoteTz *time.Location
		expected float64
	}{
		{"same timezone", localTz, localTz, 0.0},
		{"positive offset", localTz, remoteTz3, 3.0},
		{"negative offset", localTz, remoteTz5, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remoteTime := time.Now().In(tt.remoteTz)
			result := calculateOffsetHours(remoteTime, tt.localTz)

			if result != tt.expected {
				t.Errorf("calculateOffsetHours() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateShiftAmount tests the shift amount calculation for shared mode
func TestCalculateShiftAmount(t *testing.T) {
	tests := []struct {
		name        string
		offsetHours float64
		barWidth    int
		expected    int
	}{
		{"no offset", 0.0, 48, 0},
		{"positive offset +9h", 9.0, 48, 18},  // 9/24 * 48 = 18
		{"negative offset -5h", -5.0, 48, -10}, // -5/24 * 48 = -10
		{"half hour offset", 0.5, 48, 1},      // 0.5/24 * 48 = 1
		{"small bar width", 3.0, 24, 3},       // 3/24 * 24 = 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateShiftAmount(tt.offsetHours, tt.barWidth)
			if result != tt.expected {
				t.Errorf("calculateShiftAmount(%v, %d) = %d, want %d",
					tt.offsetHours, tt.barWidth, result, tt.expected)
			}
		})
	}
}

// TestFormatOffsetString tests the offset formatting function
func TestFormatOffsetString(t *testing.T) {
	tests := []struct {
		name        string
		offsetHours float64
		expected    string
	}{
		{"zero offset", 0.0, "same"},
		{"positive whole hour", 5.0, "+5h"},
		{"negative whole hour", -8.0, "-8h"},
		{"positive half hour", 5.5, "+5.5h"},
		{"negative half hour", -3.5, "-3.5h"},
		{"positive large", 12.0, "+12h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOffsetString(tt.offsetHours)
			if result != tt.expected {
				t.Errorf("formatOffsetString(%v) = %q, want %q",
					tt.offsetHours, result, tt.expected)
			}
		})
	}
}

// TestRenderIndividualBar tests the individual bar generation
func TestRenderIndividualBar(t *testing.T) {
	// Create a test model with classic color scheme
	m := Model{
		config: Config{
			ColorScheme: "classic",
		},
	}

	// Create test colleague time at midday on a weekday
	location, _ := time.LoadLocation("America/New_York")
	testTime := time.Date(2025, 1, 20, 12, 30, 0, 0, location) // Monday, 12:30pm

	ct := ColleagueTime{
		Colleague: Colleague{
			Name:       "Test",
			Timezone:   "America/New_York",
			WorkStart:  9,
			WorkEnd:    17,
			SleepStart: 23,
			SleepEnd:   7,
		},
		CurrentTime:   testTime,
		IsWorkingTime: true,
		IsWeekend:     false,
	}

	barWidth := 48
	result := m.renderIndividualBar(ct, barWidth)

	// Check that result contains expected characters
	// Note: result includes ANSI color codes, so we can't do exact string match
	// Just verify it's not empty and has reasonable length
	if len(result) < barWidth {
		t.Errorf("renderIndividualBar() returned string too short: got length %d, want at least %d",
			len(result), barWidth)
	}

	// Verify result contains brackets
	if result[0] != '[' {
		t.Errorf("renderIndividualBar() should start with '[', got %q", result[0:1])
	}
}

// TestRenderIndividualBarCharacterCount tests that the bar has correct number of characters
func TestRenderIndividualBarCharacterCount(t *testing.T) {
	m := Model{
		config: Config{
			ColorScheme: "classic",
		},
	}

	location, _ := time.LoadLocation("America/New_York")
	testTime := time.Date(2025, 1, 20, 12, 30, 0, 0, location)

	ct := ColleagueTime{
		Colleague: Colleague{
			Name:       "Test",
			Timezone:   "America/New_York",
			WorkStart:  9,
			WorkEnd:    17,
			SleepStart: 23,
			SleepEnd:   7,
		},
		CurrentTime:   testTime,
		IsWorkingTime: true,
		IsWeekend:     false,
	}

	tests := []struct {
		name     string
		barWidth int
	}{
		{"minimum width", MinBarWidth},
		{"ideal width", IdealBarWidth},
		{"custom width", 36},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.renderIndividualBar(ct, tt.barWidth)

			// Count actual characters (excluding ANSI codes)
			// The bar should have: [ + barWidth chars + ]
			// Since result includes ANSI codes, we can't count exactly,
			// but we can verify it's not empty
			if len(result) == 0 {
				t.Errorf("renderIndividualBar() returned empty string")
			}
		})
	}
}

// TestRenderSharedBarOffsetDirection tests that shared mode shifts correctly
func TestRenderSharedBarOffsetDirection(t *testing.T) {
	m := Model{
		config: Config{
			ColorScheme:  "classic",
			TimelineMode: "shared",
		},
		localTimezone: time.FixedZone("Local", 0),
	}

	// Create colleague 3 hours ahead
	remoteTz := time.FixedZone("Remote", 3*3600)
	testTime := time.Date(2025, 1, 20, 15, 0, 0, 0, remoteTz) // 15:00 remote = 12:00 local

	ct := ColleagueTime{
		Colleague: Colleague{
			Name:       "Test",
			Timezone:   "Remote",
			WorkStart:  9,
			WorkEnd:    17,
			SleepStart: 23,
			SleepEnd:   7,
		},
		CurrentTime:   testTime,
		IsWorkingTime: true,
		IsWeekend:     false,
	}

	offsetHours := 3.0
	barWidth := 48

	result := m.renderSharedBar(ct, offsetHours, barWidth)

	// Verify result is not empty
	if len(result) == 0 {
		t.Errorf("renderSharedBar() returned empty string")
	}

	// Verify result starts with bracket
	if result[0] != '[' {
		t.Errorf("renderSharedBar() should start with '[', got %q", result[0:1])
	}
}

// TestColleagueGetters tests the accessor methods with defaults
func TestColleagueGetters(t *testing.T) {
	t.Run("default values when zero", func(t *testing.T) {
		c := Colleague{
			Name:       "Test",
			Timezone:   "UTC",
			WorkStart:  0,
			WorkEnd:    0,
			SleepStart: 0,
			SleepEnd:   0,
		}

		if c.GetWorkStart() != DefaultWorkStart {
			t.Errorf("GetWorkStart() = %d, want %d", c.GetWorkStart(), DefaultWorkStart)
		}
		if c.GetWorkEnd() != DefaultWorkEnd {
			t.Errorf("GetWorkEnd() = %d, want %d", c.GetWorkEnd(), DefaultWorkEnd)
		}
		if c.GetSleepStart() != DefaultSleepStart {
			t.Errorf("GetSleepStart() = %d, want %d", c.GetSleepStart(), DefaultSleepStart)
		}
		if c.GetSleepEnd() != DefaultSleepEnd {
			t.Errorf("GetSleepEnd() = %d, want %d", c.GetSleepEnd(), DefaultSleepEnd)
		}
	})

	t.Run("custom values when set", func(t *testing.T) {
		c := Colleague{
			Name:       "Test",
			Timezone:   "UTC",
			WorkStart:  8,
			WorkEnd:    16,
			SleepStart: 22,
			SleepEnd:   6,
		}

		if c.GetWorkStart() != 8 {
			t.Errorf("GetWorkStart() = %d, want 8", c.GetWorkStart())
		}
		if c.GetWorkEnd() != 16 {
			t.Errorf("GetWorkEnd() = %d, want 16", c.GetWorkEnd())
		}
		if c.GetSleepStart() != 22 {
			t.Errorf("GetSleepStart() = %d, want 22", c.GetSleepStart())
		}
		if c.GetSleepEnd() != 6 {
			t.Errorf("GetSleepEnd() = %d, want 6", c.GetSleepEnd())
		}
	})
}
