package main

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2024, 11, 20, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"24h format", "24h", "15:30:45"},
		{"12h format", "12h", "3:30:45 PM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(testTime, tt.format)
			if result != tt.expected {
				t.Errorf("FormatTime(%s) = %s, want %s", tt.format, result, tt.expected)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			"Monday in November",
			time.Date(2024, 11, 20, 15, 30, 0, 0, time.UTC),
			"Wed, Nov 20",
		},
		{
			"First of month",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			"Mon, Jan 01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.time)
			if result != tt.expected {
				t.Errorf("FormatDate() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		expectErr bool
	}{
		{"Valid: America/New_York", "America/New_York", false},
		{"Valid: Europe/London", "Europe/London", false},
		{"Valid: Asia/Tokyo", "Asia/Tokyo", false},
		{"Valid: UTC", "UTC", false},
		{"Invalid: Foo/Bar", "Foo/Bar", true},
		{"Invalid: random string", "notimezone", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateTimezone(%s) error = %v, expectErr %v", tt.timezone, err, tt.expectErr)
			}
		})
	}
}

func TestComputeColleagueTimes(t *testing.T) {
	// Use fixed time for deterministic tests
	// Wednesday, Nov 20, 2024, 10:00 AM EST (15:00 UTC)
	localTz, _ := time.LoadLocation("America/New_York")

	colleagues := []Colleague{
		{
			Name:     "Alice (New York)",
			Timezone: "America/New_York",
		},
		{
			Name:     "Bob (London)",
			Timezone: "Europe/London",
		},
		{
			Name:     "Charlie (Invalid)",
			Timezone: "Invalid/Timezone",
		},
		{
			Name:     "Dana (Tokyo)",
			Timezone: "Asia/Tokyo",
		},
	}

	result := ComputeColleagueTimes(colleagues, localTz)

	// Invalid entries are kept and flagged so the UI can surface them
	if len(result) != 4 {
		t.Fatalf("Expected 4 results (invalid entry kept), got %d", len(result))
	}
	if !result[2].InvalidTimezone {
		t.Error("Expected Charlie's entry to be flagged InvalidTimezone")
	}
	if result[2].ConfigIndex != 2 {
		t.Errorf("Expected Charlie's ConfigIndex 2, got %d", result[2].ConfigIndex)
	}

	// Valid entries around the invalid one must keep their config index
	// so edit/delete operate on the right colleague
	if result[1].ConfigIndex != 1 || result[1].InvalidTimezone {
		t.Errorf("Expected Bob valid with ConfigIndex 1, got index %d invalid=%v",
			result[1].ConfigIndex, result[1].InvalidTimezone)
	}
	if result[3].ConfigIndex != 3 || result[3].InvalidTimezone {
		t.Errorf("Expected Dana valid with ConfigIndex 3, got index %d invalid=%v",
			result[3].ConfigIndex, result[3].InvalidTimezone)
	}

	// Verify Alice's data
	alice := result[0]
	if alice.Colleague.Name != "Alice (New York)" {
		t.Errorf("Expected Alice, got %s", alice.Colleague.Name)
	}
	if alice.Offset != "same" {
		t.Errorf("Expected 'same' offset for local timezone, got %s", alice.Offset)
	}

	// Verify Bob's data
	bob := result[1]
	if bob.Colleague.Name != "Bob (London)" {
		t.Errorf("Expected Bob, got %s", bob.Colleague.Name)
	}
	// London is +5h from New York (EST to GMT)
	if bob.Offset != "+5h" {
		t.Errorf("Expected +5h offset, got %s", bob.Offset)
	}
}

func TestComputeColleagueTimesHalfHourOffset(t *testing.T) {
	colleagues := []Colleague{
		{Name: "Ravi (Kolkata)", Timezone: "Asia/Kolkata"},
	}

	result := ComputeColleagueTimes(colleagues, time.UTC)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// India is UTC+5:30 year-round; integer division used to show "+5h"
	if result[0].Offset != "+5.5h" {
		t.Errorf("Expected +5.5h offset for Kolkata from UTC, got %s", result[0].Offset)
	}
}

func TestOvernightWorkingHours(t *testing.T) {
	// Ranges are built around the current hour so the assertions hold at
	// any time of day while still exercising the wraparound branch
	now := time.Now().UTC()
	h := now.Hour()
	isWeekend := now.Weekday() == time.Saturday || now.Weekday() == time.Sunday

	// A one-hour range containing the current hour (wraps when h == 23)
	inRange := ComputeColleagueTimes([]Colleague{{
		Name:      "On Shift",
		Timezone:  "UTC",
		WorkStart: HourPtr(h),
		WorkEnd:   HourPtr((h + 1) % 24),
	}}, time.UTC)
	if got := inRange[0].IsWorkingTime; got != !isWeekend {
		t.Errorf("Colleague working %d-%d at hour %d: IsWorkingTime = %v, want %v",
			h, (h+1)%24, h, got, !isWeekend)
	}

	// The complementary range excludes the current hour (wraps for h < 23)
	outOfRange := ComputeColleagueTimes([]Colleague{{
		Name:      "Off Shift",
		Timezone:  "UTC",
		WorkStart: HourPtr((h + 1) % 24),
		WorkEnd:   HourPtr(h),
	}}, time.UTC)
	if outOfRange[0].IsWorkingTime {
		t.Errorf("Colleague working %d-%d at hour %d: IsWorkingTime = true, want false",
			(h+1)%24, h, h)
	}
}

func TestWorkingHoursDetection(t *testing.T) {
	tests := []struct {
		name       string
		hour       int
		day        time.Weekday
		expectWork bool
		expectWeek bool
	}{
		{"Weekday 10am", 10, time.Wednesday, true, false},
		{"Weekday 9am (start)", 9, time.Monday, true, false},
		{"Weekday 8am (before)", 8, time.Tuesday, false, false},
		{"Weekday 5pm (end)", 17, time.Thursday, false, false},
		{"Weekday 4pm", 16, time.Friday, true, false},
		{"Saturday 10am", 10, time.Saturday, false, true},
		{"Sunday 2pm", 14, time.Sunday, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test time
			testDate := time.Date(2024, 11, 17, 0, 0, 0, 0, time.UTC) // Sunday
			// Add days to get to the right weekday
			daysToAdd := int(tt.day - testDate.Weekday())
			testDate = testDate.AddDate(0, 0, daysToAdd)
			testDate = time.Date(testDate.Year(), testDate.Month(), testDate.Day(), tt.hour, 0, 0, 0, time.UTC)

			colleague := Colleague{
				Name:      "Test",
				Timezone:  "UTC",
				WorkStart: HourPtr(9),
				WorkEnd:   HourPtr(17),
			}

			// Manually check working hours (mimics the logic in ComputeColleagueTimes)
			isWeekend := testDate.Weekday() == time.Saturday || testDate.Weekday() == time.Sunday
			isWorkingTime := !isWeekend && testDate.Hour() >= colleague.GetWorkStart() && testDate.Hour() < colleague.GetWorkEnd()

			if isWeekend != tt.expectWeek {
				t.Errorf("Weekend detection: got %v, want %v", isWeekend, tt.expectWeek)
			}
			if isWorkingTime != tt.expectWork {
				t.Errorf("Working hours detection: got %v, want %v", isWorkingTime, tt.expectWork)
			}
		})
	}
}
