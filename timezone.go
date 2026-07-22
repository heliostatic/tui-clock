package main

import (
	"fmt"
	"time"
)

// ComputeColleagueTimes calculates current time and metadata for all
// colleagues; entries whose timezone fails to load are kept in the
// list flagged InvalidTimezone so the user can see, fix, or delete
// them in the UI
func ComputeColleagueTimes(colleagues []Colleague, localTz *time.Location) []ColleagueTime {
	now := time.Now()
	localNow := now.In(localTz)

	result := make([]ColleagueTime, 0, len(colleagues))

	for i, colleague := range colleagues {
		loc, err := time.LoadLocation(colleague.Timezone)
		if err != nil {
			result = append(result, ColleagueTime{
				Colleague:       colleague,
				ConfigIndex:     i,
				InvalidTimezone: true,
			})
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

		// Check if it's working time (accessors supply defaults for unset
		// hours; isInTimeRange handles overnight ranges like 16-0)
		hour := colleagueTime.Hour()
		isWorkingTime := !isWeekend && isInTimeRange(hour, colleague.GetWorkStart(), colleague.GetWorkEnd())

		// Surface upcoming DST transitions so offset changes don't surprise
		dstAt, dstDelta, hasDST := nextOffsetChange(loc, now, DSTLookahead)

		result = append(result, ColleagueTime{
			Colleague:     colleague,
			ConfigIndex:   i,
			CurrentTime:   colleagueTime,
			Offset:        offsetStr,
			IsWorkingTime: isWorkingTime,
			IsWeekend:     isWeekend,
			DSTChangeAt:   dstAt,
			DSTDeltaHours: dstDelta,
			HasDSTChange:  hasDST,
		})
	}

	return result
}

// DSTLookahead is how far ahead colleagues' upcoming UTC-offset
// changes (DST transitions) are surfaced in the list view
const DSTLookahead = 7 * 24 * time.Hour

// nextOffsetChange finds the next moment loc's UTC offset changes
// within lookahead of from. Returns the transition time (in loc), the
// offset delta in hours, and whether a change was found. A window
// containing two transitions that cancel out is treated as no change;
// real zones never transition twice in a week.
func nextOffsetChange(loc *time.Location, from time.Time, lookahead time.Duration) (time.Time, float64, bool) {
	_, startOff := from.In(loc).Zone()
	end := from.Add(lookahead)
	_, endOff := end.In(loc).Zone()
	if startOff == endOff {
		return time.Time{}, 0, false
	}

	// Binary search the boundary to minute precision
	lo, hi := from, end
	for hi.Sub(lo) > time.Minute {
		mid := lo.Add(hi.Sub(lo) / 2)
		if _, off := mid.In(loc).Zone(); off == startOff {
			lo = mid
		} else {
			hi = mid
		}
	}

	_, afterOff := hi.In(loc).Zone()
	return hi.In(loc), float64(afterOff-startOff) / 3600.0, true
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
