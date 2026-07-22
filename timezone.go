package main

import (
	"fmt"
	"time"
)

// ComputeColleagueTimes calculates current time and metadata for all colleagues
func ComputeColleagueTimes(colleagues []Colleague, localTz *time.Location, timeFormat string) ([]ColleagueTime, error) {
	now := time.Now()
	localNow := now.In(localTz)

	result := make([]ColleagueTime, 0, len(colleagues))

	for i, colleague := range colleagues {
		loc, err := time.LoadLocation(colleague.Timezone)
		if err != nil {
			// Skip invalid timezones but continue processing others
			continue
		}

		colleagueTime := now.In(loc)

		// Calculate offset in fractional hours so half-hour zones
		// (e.g. India +5:30) display correctly
		_, localOffset := localNow.Zone()
		_, colleagueOffset := colleagueTime.Zone()
		offsetHours := float64(colleagueOffset-localOffset) / 3600.0
		offsetStr := formatOffsetString(offsetHours)

		// Check if it's weekend
		isWeekend := colleagueTime.Weekday() == time.Saturday || colleagueTime.Weekday() == time.Sunday

		// Check if it's working time (accessors supply defaults for unset hours)
		hour := colleagueTime.Hour()
		isWorkingTime := !isWeekend && hour >= colleague.GetWorkStart() && hour < colleague.GetWorkEnd()

		result = append(result, ColleagueTime{
			Colleague:     colleague,
			ConfigIndex:   i,
			CurrentTime:   colleagueTime,
			Offset:        offsetStr,
			IsWorkingTime: isWorkingTime,
			IsWeekend:     isWeekend,
		})
	}

	return result, nil
}

// FormatTime formats a time according to the specified format
func FormatTime(t time.Time, format string) string {
	if format == "12h" {
		return t.Format("3:04:05 PM")
	}
	return t.Format("15:04:05")
}

// FormatDate formats the date and day of week
func FormatDate(t time.Time) string {
	return t.Format("Mon, Jan 02")
}

// ValidateTimezone checks if a timezone string is valid
func ValidateTimezone(tz string) error {
	_, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone '%s': %w", tz, err)
	}
	return nil
}
