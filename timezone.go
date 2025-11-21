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

	for _, colleague := range colleagues {
		loc, err := time.LoadLocation(colleague.Timezone)
		if err != nil {
			// Skip invalid timezones but continue processing others
			continue
		}

		colleagueTime := now.In(loc)

		// Calculate offset
		_, localOffset := localNow.Zone()
		_, colleagueOffset := colleagueTime.Zone()
		offsetHours := (colleagueOffset - localOffset) / 3600

		var offsetStr string
		if offsetHours == 0 {
			offsetStr = "same"
		} else if offsetHours > 0 {
			offsetStr = fmt.Sprintf("+%dh", offsetHours)
		} else {
			offsetStr = fmt.Sprintf("%dh", offsetHours)
		}

		// Check if it's weekend
		isWeekend := colleagueTime.Weekday() == time.Saturday || colleagueTime.Weekday() == time.Sunday

		// Check if it's working time
		hour := colleagueTime.Hour()
		isWorkingTime := !isWeekend && hour >= colleague.WorkStart && hour < colleague.WorkEnd

		result = append(result, ColleagueTime{
			Colleague:     colleague,
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
