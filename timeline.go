package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Timeline rendering functions for the world clock application.
// Provides two visualization modes:
// - Individual: Each colleague has their own 24-hour timeline
// - Shared: Single timeline showing all colleagues shifted by offset

// renderTimeline renders the complete timeline view
func (m Model) renderTimeline() string {
	var b strings.Builder

	// Header (displayNow applies any scrub offset)
	localTime := m.displayNow()
	header := fmt.Sprintf("🌍 Timeline View - Local Time: %s (%s)",
		FormatTime(localTime, m.config.TimeFormat),
		FormatDate(localTime))
	if m.timeOffset != 0 {
		header += fmt.Sprintf("  ⏩ scrubbed %s", formatOffsetString(m.timeOffset.Hours()))
	}
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Calculate visible range
	start := m.scrollOffset
	end := min(start+MaxVisible, len(m.colleagues))

	// Show scroll indicators
	topIndicator, bottomIndicator := renderScrollIndicators(m.scrollOffset, MaxVisible, len(m.colleagues))
	b.WriteString(topIndicator)

	// Render visible colleagues (shifted by any scrub offset)
	for i := start; i < end; i++ {
		ct := m.scrubbed(m.colleagues[i])
		switch {
		case ct.InvalidTimezone:
			b.WriteString(m.renderInvalidTimelineRow(ct))
		case m.config.TimelineMode == "individual":
			b.WriteString(m.renderTimelineRow(i, ct))
		default:
			b.WriteString(m.renderSharedTimelineRow(i, ct))
		}
		b.WriteString("\n")
	}

	// Show bottom scroll indicator
	b.WriteString(bottomIndicator)

	// Team overlap summary (shared mode, two or more valid colleagues)
	if m.config.TimelineMode == "shared" {
		if row := m.renderOverlapRow(); row != "" {
			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	// Add hour labels once at bottom for both modes
	barWidth := m.calculateTimelineBarWidth()
	labels := m.renderHourLabels(barWidth, NameFieldWidth+TimeFieldWidth+2)
	b.WriteString(labels)
	b.WriteString("\n")

	// Legend
	b.WriteString(m.renderTimelineLegend())
	b.WriteString("\n")

	// Footer with keybindings
	b.WriteString(m.renderTimelineFooter())

	return b.String()
}

// renderTimelineRow renders a single colleague's timeline row (without labels)
func (m Model) renderTimelineRow(index int, ct ColleagueTime) string {
	// Name and location (max NameFieldWidth chars)
	nameStr := ct.Colleague.Name
	nameStr = truncateOrPad(nameStr, NameFieldWidth)

	// Current time
	timeStr := FormatTime(ct.CurrentTime, m.config.TimeFormat)
	timeStr = truncateOrPad(timeStr, TimeFieldWidth)

	// Calculate bar width
	barWidth := m.calculateTimelineBarWidth()

	// Generate the timeline bar
	bar := m.renderIndividualBar(ct, barWidth)

	// Apply name styling based on current status
	nameStyle := getNameStyle(ct)

	return fmt.Sprintf("%s %s %s",
		nameStyle.Render(nameStr),
		timeStr,
		bar)
}

// barCharForHour classifies an hour of a colleague's day into a bar
// character. Configured work hours take precedence over sleep hours:
// a night-shift colleague working 0-8 should render as working even
// though the default sleep range (23-7) overlaps those hours.
func barCharForHour(ct ColleagueTime, hour int) rune {
	if !ct.IsWeekend && isInTimeRange(hour, ct.Colleague.GetWorkStart(), ct.Colleague.GetWorkEnd()) {
		return '█' // Work hours
	}
	if isInTimeRange(hour, ct.Colleague.GetSleepStart(), ct.Colleague.GetSleepEnd()) {
		return '░' // Sleep
	}
	return '▓' // Awake off-hours
}

// renderIndividualBar generates a timeline bar for individual mode
func (m Model) renderIndividualBar(ct ColleagueTime, barWidth int) string {
	bar := make([]rune, barWidth)

	// Get current hour position (0-24)
	currentHour := float64(ct.CurrentTime.Hour())
	currentMinute := float64(ct.CurrentTime.Minute())
	currentPosition := (currentHour + currentMinute/60.0) / 24.0 // 0.0 to 1.0
	markerIndex := int(currentPosition * float64(barWidth))

	// Build bar character by character
	for i := range barWidth {
		// Calculate which hour(s) this position represents
		hourFraction := float64(i) / float64(barWidth)
		hour := int(hourFraction * 24.0)

		// Marker position will be colored differently, not replaced
		bar[i] = barCharForHour(ct, hour)
	}

	// Apply colors
	return m.colorizeBar(bar, ct, markerIndex)
}

// colorizeBar applies color styling to the timeline bar
func (m Model) colorizeBar(bar []rune, ct ColleagueTime, markerIndex int) string {
	scheme := getCurrentColorScheme(m.config.ColorScheme)
	var result strings.Builder

	result.WriteString("[")

	for i, char := range bar {
		var style lipgloss.Style

		if i == markerIndex {
			// Current time marker
			style = lipgloss.NewStyle().
				Foreground(scheme.MarkerColor).
				Bold(true)
		} else {
			switch char {
			case '░': // Sleep
				style = lipgloss.NewStyle().Foreground(scheme.SleepColor)
			case '▓': // Awake off
				style = lipgloss.NewStyle().Foreground(scheme.AwakeOffColor)
			case '█': // Work
				if ct.IsWeekend {
					style = lipgloss.NewStyle().Foreground(scheme.WeekendTint)
				} else {
					style = lipgloss.NewStyle().Foreground(scheme.WorkColor)
				}
			}
		}

		result.WriteString(style.Render(string(char)))
	}

	result.WriteString("]")
	return result.String()
}

// renderHourLabels renders the hour labels below the timeline bar
func (m Model) renderHourLabels(barWidth int, leftPadding int) string {
	// Build the label line character by character for proper alignment
	labelChars := make([]rune, barWidth+2) // +2 for brackets

	// Initialize with spaces
	for i := range labelChars {
		labelChars[i] = ' '
	}

	// Set brackets
	labelChars[0] = '['
	labelChars[barWidth+1] = ']'

	// Calculate positions for each hour (0, 6, 12, 18, 24)
	// Hour X should be at position: (X / 24.0) * barWidth
	hours := []int{0, 6, 12, 18, 24}

	for _, hour := range hours {
		// Calculate exact position for this hour in the bar
		centerPos := int(float64(hour) / 24.0 * float64(barWidth))

		// Convert hour to string
		label := fmt.Sprintf("%d", hour)
		labelLen := len(label)

		// Calculate start position to center the label
		// +1 accounts for the opening bracket
		startPos := max(
			// Ensure we don't go out of bounds
			centerPos-(labelLen/2)+1, 1)
		if startPos+labelLen > barWidth+1 {
			startPos = barWidth + 1 - labelLen
		}

		// Place each character of the label
		for i, ch := range label {
			pos := startPos + i
			if pos > 0 && pos <= barWidth {
				labelChars[pos] = ch
			}
		}
	}

	// Add left padding
	padding := strings.Repeat(" ", leftPadding)

	return footerStyle.Render(padding + string(labelChars))
}

// renderTimelineLegend renders the legend explaining timeline symbols
func (m Model) renderTimelineLegend() string {
	scheme := getCurrentColorScheme(m.config.ColorScheme)

	sleep := lipgloss.NewStyle().Foreground(scheme.SleepColor).Render("░")
	awake := lipgloss.NewStyle().Foreground(scheme.AwakeOffColor).Render("▓")
	work := lipgloss.NewStyle().Foreground(scheme.WorkColor).Render("█")
	// Show the marker as a highlighted block to indicate color highlighting
	marker := lipgloss.NewStyle().Foreground(scheme.MarkerColor).Bold(true).Render("█")

	legend := fmt.Sprintf("\n%s sleep • %s off-hours • %s work • %s now",
		sleep, awake, work, marker)

	// Overlap row legend (shared mode only)
	if m.config.TimelineMode == "shared" {
		all := lipgloss.NewStyle().Foreground(scheme.Success).Render("█")
		some := lipgloss.NewStyle().Foreground(scheme.Warning).Render("▓")
		legend += fmt.Sprintf("\noverlap: %s everyone working • %s majority", all, some)
	}

	return footerStyle.Render(legend)
}

// renderTimelineFooter renders the footer with keybindings for timeline mode
func (m Model) renderTimelineFooter() string {
	mode := "individual"
	if m.config.TimelineMode == "shared" {
		mode = "shared"
	}

	help := []string{
		"t normal mode",
		"m " + mode,
		"↑/↓ scroll",
		"←/→ scrub time",
		"c cycle colors",
		"? help",
		"q quit",
	}
	if m.timeOffset != 0 {
		help = append(help, "esc back to now")
	}
	return footerStyle.Render(strings.Join(help, " • "))
}

// calculateTimelineBarWidth calculates the appropriate bar width based on terminal size
func (m Model) calculateTimelineBarWidth() int {
	// Reserve space for name, time, padding, and brackets
	reservedSpace := NameFieldWidth + TimeFieldWidth + 5 + 2
	available := m.width - reservedSpace

	// Ensure minimum
	if available < MinBarWidth {
		return MinBarWidth
	}

	// Cap at ideal width
	if available > IdealBarWidth {
		return IdealBarWidth
	}

	// Return available space
	return available
}

// isInTimeRange checks if an hour falls within a time range (handles
// wraparound like 23-7). The range is half-open [start, end), so
// start == end is an empty range that never matches.
func isInTimeRange(hour, start, end int) bool {
	if start <= end {
		// Normal range (e.g., 9-17)
		return hour >= start && hour < end
	}
	// Wraparound range (e.g., 23-7)
	return hour >= start || hour < end
}

// truncateOrPad truncates or pads a string to an exact display width,
// measured in terminal cells (not bytes) so non-ASCII names stay aligned
func truncateOrPad(s string, width int) string {
	return runewidth.FillRight(runewidth.Truncate(s, width, ""), width)
}

// calculateOffsetHours calculates the hour offset between two times
func calculateOffsetHours(t time.Time, localTz *time.Location) float64 {
	localTime := time.Now().In(localTz)
	_, localOffset := localTime.Zone()
	_, remoteOffset := t.Zone()

	offsetSeconds := remoteOffset - localOffset
	return float64(offsetSeconds) / 3600.0
}

// renderInvalidTimelineRow renders a warning row for a colleague whose
// timezone failed to load (no time or bar can be computed)
func (m Model) renderInvalidTimelineRow(ct ColleagueTime) string {
	nameStr := truncateOrPad(ct.Colleague.Name, NameFieldWidth)
	msg := fmt.Sprintf("⚠ invalid timezone %q — edit or delete", ct.Colleague.Timezone)
	return fmt.Sprintf("%s %s", invalidStyle.Render(nameStr), invalidStyle.Render(msg))
}

// renderSharedTimelineRow renders a single colleague's row in shared timeline mode
func (m Model) renderSharedTimelineRow(index int, ct ColleagueTime) string {
	// Calculate offset hours
	offsetHours := calculateOffsetHours(ct.CurrentTime, m.localTimezone)

	// Name (same format as individual mode)
	nameStr := ct.Colleague.Name
	nameStr = truncateOrPad(nameStr, NameFieldWidth)

	// Current time (same format as individual mode)
	timeStr := FormatTime(ct.CurrentTime, m.config.TimeFormat)
	timeStr = truncateOrPad(timeStr, TimeFieldWidth)

	// Calculate bar width
	barWidth := m.calculateTimelineBarWidth()

	// Generate shifted timeline bar
	bar := m.renderSharedBar(ct, offsetHours, barWidth)

	// Apply name styling based on current status
	nameStyle := getNameStyle(ct)

	return fmt.Sprintf("%s %s %s", nameStyle.Render(nameStr), timeStr, bar)
}

// renderSharedBar generates a timeline bar for shared mode (shifted by offset)
func (m Model) renderSharedBar(ct ColleagueTime, offsetHours float64, barWidth int) string {
	bar := make([]rune, barWidth)

	// Calculate current time marker position (local time, scrub-aware)
	localTime := m.displayNow()
	currentHour := float64(localTime.Hour())
	currentMinute := float64(localTime.Minute())
	currentPosition := (currentHour + currentMinute/60.0) / 24.0
	markerIndex := int(currentPosition * float64(barWidth))

	// Build bar with shift applied
	for i := range barWidth {
		bar[i] = barCharForHour(ct, sharedBarHour(i, barWidth, offsetHours))
	}

	// Pass markerIndex to colorizer for proper styling (will highlight that position)
	return m.colorizeBar(bar, ct, markerIndex)
}

// sharedBarHour converts a shared-mode bar position (which represents
// local time) into the colleague's local hour by applying their offset,
// wrapping at day boundaries. If they're +3h ahead, when local is
// 12:00 they're at 15:00.
func sharedBarHour(position, barWidth int, offsetHours float64) int {
	theirHourFraction := float64(position)/float64(barWidth) + offsetHours/24.0

	// Wrap around if needed
	for theirHourFraction < 0 {
		theirHourFraction += 1.0
	}
	for theirHourFraction >= 1.0 {
		theirHourFraction -= 1.0
	}

	return int(theirHourFraction * 24.0)
}

// computeSharedOverlap returns, for each shared-bar position, how many
// of the given colleagues are working at that moment of the local day,
// plus the number of colleagues counted. Invalid-timezone entries are
// ignored.
func computeSharedOverlap(colleagues []ColleagueTime, localTz *time.Location, barWidth int) ([]int, int) {
	counts := make([]int, barWidth)
	total := 0

	for _, ct := range colleagues {
		if ct.InvalidTimezone {
			continue
		}
		total++
		offsetHours := calculateOffsetHours(ct.CurrentTime, localTz)
		for i := range counts {
			if barCharForHour(ct, sharedBarHour(i, barWidth, offsetHours)) == '█' {
				counts[i]++
			}
		}
	}

	return counts, total
}

// renderOverlapRow renders the team-overlap summary row for shared
// mode: where everyone is working, where a majority is, and how many
// are working right now. Returns "" when fewer than two colleagues
// have valid timezones.
func (m Model) renderOverlapRow() string {
	barWidth := m.calculateTimelineBarWidth()

	// Count against scrubbed times so the row follows time scrubbing
	// (the weekday, and with it the work blocks, can change)
	cts := make([]ColleagueTime, len(m.colleagues))
	for i, ct := range m.colleagues {
		cts[i] = m.scrubbed(ct)
	}
	counts, total := computeSharedOverlap(cts, m.localTimezone, barWidth)
	if total < 2 {
		return ""
	}

	// Current time marker at the local-time position, like every shared row
	localTime := m.displayNow()
	currentPosition := (float64(localTime.Hour()) + float64(localTime.Minute())/60.0) / 24.0
	markerIndex := int(currentPosition * float64(barWidth))

	scheme := getCurrentColorScheme(m.config.ColorScheme)
	allStyle := lipgloss.NewStyle().Foreground(scheme.Success)
	someStyle := lipgloss.NewStyle().Foreground(scheme.Warning)
	noneStyle := lipgloss.NewStyle().Foreground(scheme.Muted)
	markerStyle := lipgloss.NewStyle().Foreground(scheme.MarkerColor).Bold(true)

	var bar strings.Builder
	bar.WriteString("[")
	for i, count := range counts {
		var char string
		var style lipgloss.Style
		switch {
		case count == total:
			char, style = "█", allStyle
		case count*2 >= total:
			char, style = "▓", someStyle
		default:
			char, style = "░", noneStyle
		}
		if i == markerIndex {
			style = markerStyle
		}
		bar.WriteString(style.Render(char))
	}
	bar.WriteString("]")

	nameStr := truncateOrPad("Team overlap", NameFieldWidth)
	nowStr := truncateOrPad(fmt.Sprintf("%d/%d now", counts[markerIndex], total), TimeFieldWidth)

	// offHoursStyle: muted like the footer but without its top margin
	return fmt.Sprintf("%s %s %s", offHoursStyle.Render(nameStr), nowStr, bar.String())
}

// formatOffsetString formats the offset hours as a string
func formatOffsetString(offsetHours float64) string {
	if offsetHours == 0 {
		return "same"
	}

	sign := "+"
	if offsetHours < 0 {
		sign = "-"
		offsetHours = -offsetHours
	}

	// Shortest exact decimal: real-world offsets are quarter-hour
	// multiples, so this yields "+5h", "+5.5h", "+5.75h" — never a
	// rounded value like "+5.8h" for Nepal's +5:45
	return sign + strconv.FormatFloat(offsetHours, 'f', -1, 64) + "h"
}
