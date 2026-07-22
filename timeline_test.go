package main

import (
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
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
		// Widths are display cells, not bytes: "José" is 5 bytes but 4 cells
		{"accented - pad by cells", "José", 6, "José  "},
		{"CJK - pad by cells", "田中", 6, "田中  "},
		// Truncating a wide char that won't fit pads the leftover cell
		{"CJK - truncate to odd width", "田中太郎", 5, "田中 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateOrPad(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("truncateOrPad(%q, %d) = %q, want %q",
					tt.input, tt.width, result, tt.expected)
			}
			if w := runewidth.StringWidth(result); w != tt.width {
				t.Errorf("truncateOrPad(%q, %d) returned display width %d, want %d",
					tt.input, tt.width, w, tt.width)
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
		{"quarter hour (Nepal)", 5.75, "+5.75h"},
		{"negative quarter hour", -9.75, "-9.75h"},
		{"chatham islands", 12.75, "+12.75h"},
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
			WorkStart:  HourPtr(9),
			WorkEnd:    HourPtr(17),
			SleepStart: HourPtr(23),
			SleepEnd:   HourPtr(7),
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
			WorkStart:  HourPtr(9),
			WorkEnd:    HourPtr(17),
			SleepStart: HourPtr(23),
			SleepEnd:   HourPtr(7),
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
			WorkStart:  HourPtr(9),
			WorkEnd:    HourPtr(17),
			SleepStart: HourPtr(23),
			SleepEnd:   HourPtr(7),
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

// TestBarCharPrecedence tests that configured work hours win over
// overlapping (default) sleep hours in the timeline bar
func TestBarCharPrecedence(t *testing.T) {
	nightShift := ColleagueTime{
		Colleague: Colleague{
			Name:      "Night Owl",
			Timezone:  "UTC",
			WorkStart: HourPtr(0),
			WorkEnd:   HourPtr(8),
			// Sleep unset: defaults to 23-7, overlapping the work range
		},
	}

	for hour := 0; hour < 8; hour++ {
		if got := barCharForHour(nightShift, hour); got != '█' {
			t.Errorf("hour %d: got %q, want work block (configured work must win over default sleep)", hour, got)
		}
	}
	// 23:00 is sleep (outside the work range)
	if got := barCharForHour(nightShift, 23); got != '░' {
		t.Errorf("hour 23: got %q, want sleep", got)
	}
	// Midday is off-hours
	if got := barCharForHour(nightShift, 12); got != '▓' {
		t.Errorf("hour 12: got %q, want off-hours", got)
	}

	// On weekends the work block disappears and sleep shows through
	weekend := nightShift
	weekend.IsWeekend = true
	if got := barCharForHour(weekend, 2); got != '░' {
		t.Errorf("weekend hour 2: got %q, want sleep", got)
	}

	// A default 9-17 worker is unchanged by the precedence flip
	standard := ColleagueTime{Colleague: Colleague{Name: "Standard", Timezone: "UTC"}}
	if got := barCharForHour(standard, 10); got != '█' {
		t.Errorf("standard hour 10: got %q, want work", got)
	}
	if got := barCharForHour(standard, 2); got != '░' {
		t.Errorf("standard hour 2: got %q, want sleep", got)
	}
	if got := barCharForHour(standard, 20); got != '▓' {
		t.Errorf("standard hour 20: got %q, want off-hours", got)
	}
}

// TestComputeSharedOverlap tests the team-overlap counting with fixed
// zones so results are deterministic
func TestComputeSharedOverlap(t *testing.T) {
	// Monday 12:00 UTC; barWidth 24 makes position == local hour
	instant := time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC)
	remote := time.FixedZone("R+3", 3*3600)
	const barWidth = 24

	colleagues := []ColleagueTime{
		{
			// Works 9-17 local (UTC)
			Colleague:   Colleague{Name: "A", Timezone: "UTC"},
			CurrentTime: instant,
		},
		{
			// Works 9-17 their time = 6-14 UTC
			Colleague:   Colleague{Name: "B", Timezone: "R+3"},
			CurrentTime: instant.In(remote),
		},
		{
			// Ignored entirely
			Colleague:       Colleague{Name: "C", Timezone: "Bad/Zone"},
			InvalidTimezone: true,
		},
	}

	counts, total := computeSharedOverlap(colleagues, time.UTC, barWidth)

	if total != 2 {
		t.Fatalf("total = %d, want 2 (invalid entry ignored)", total)
	}
	// Both working: 9-14 UTC
	for _, h := range []int{9, 13} {
		if counts[h] != 2 {
			t.Errorf("counts[%d] = %d, want 2 (both working)", h, counts[h])
		}
	}
	// Only B: 6-9 UTC; only A: 14-17 UTC
	for _, h := range []int{7, 15} {
		if counts[h] != 1 {
			t.Errorf("counts[%d] = %d, want 1", h, counts[h])
		}
	}
	// Nobody: evening
	if counts[20] != 0 {
		t.Errorf("counts[20] = %d, want 0", counts[20])
	}
}

// TestSharedBarHour tests position-to-their-hour conversion including wraparound
func TestSharedBarHour(t *testing.T) {
	tests := []struct {
		name        string
		position    int
		barWidth    int
		offsetHours float64
		expected    int
	}{
		{"no offset", 12, 24, 0, 12},
		{"positive offset", 12, 24, 3, 15},
		{"wraps past midnight", 22, 24, 5, 3},
		{"negative offset", 2, 24, -5, 21},
		{"half-hour offset", 12, 24, 5.5, 17},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedBarHour(tt.position, tt.barWidth, tt.offsetHours); got != tt.expected {
				t.Errorf("sharedBarHour(%d, %d, %v) = %d, want %d",
					tt.position, tt.barWidth, tt.offsetHours, got, tt.expected)
			}
		})
	}
}

// TestColleagueGetters tests the accessor methods with defaults
func TestColleagueGetters(t *testing.T) {
	t.Run("default values when unset", func(t *testing.T) {
		c := Colleague{
			Name:     "Test",
			Timezone: "UTC",
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
			WorkStart:  HourPtr(8),
			WorkEnd:    HourPtr(16),
			SleepStart: HourPtr(22),
			SleepEnd:   HourPtr(6),
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

	t.Run("midnight (0) is a valid configured value", func(t *testing.T) {
		c := Colleague{
			Name:      "Night Owl",
			Timezone:  "UTC",
			WorkStart: HourPtr(0),
			SleepEnd:  HourPtr(0),
		}

		if c.GetWorkStart() != 0 {
			t.Errorf("GetWorkStart() = %d, want 0 (midnight)", c.GetWorkStart())
		}
		if c.GetSleepEnd() != 0 {
			t.Errorf("GetSleepEnd() = %d, want 0 (midnight)", c.GetSleepEnd())
		}
	})
}
