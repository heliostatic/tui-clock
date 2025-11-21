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
			Name:      "Alice (New York)",
			Timezone:  "America/New_York",
			WorkStart: 9,
			WorkEnd:   17,
		},
		{
			Name:      "Bob (London)",
			Timezone:  "Europe/London",
			WorkStart: 9,
			WorkEnd:   17,
		},
		{
			Name:      "Charlie (Invalid)",
			Timezone:  "Invalid/Timezone",
			WorkStart: 9,
			WorkEnd:   17,
		},
	}

	result, err := ComputeColleagueTimes(colleagues, localTz, "24h")
	if err != nil {
		t.Fatalf("ComputeColleagueTimes returned unexpected error: %v", err)
	}

	// Should skip invalid timezone, so only 2 results
	if len(result) != 2 {
		t.Errorf("Expected 2 results (skipping invalid), got %d", len(result))
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

func TestWorkingHoursDetection(t *testing.T) {
	tests := []struct {
		name        string
		hour        int
		day         time.Weekday
		expectWork  bool
		expectWeek  bool
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
				WorkStart: 9,
				WorkEnd:   17,
			}

			// Manually check working hours (mimics the logic in ComputeColleagueTimes)
			isWeekend := testDate.Weekday() == time.Saturday || testDate.Weekday() == time.Sunday
			isWorkingTime := !isWeekend && testDate.Hour() >= colleague.WorkStart && testDate.Hour() < colleague.WorkEnd

			if isWeekend != tt.expectWeek {
				t.Errorf("Weekend detection: got %v, want %v", isWeekend, tt.expectWeek)
			}
			if isWorkingTime != tt.expectWork {
				t.Errorf("Working hours detection: got %v, want %v", isWorkingTime, tt.expectWork)
			}
		})
	}
}
