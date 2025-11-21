package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Timeline rendering functions for the world clock application.
// Provides two visualization modes:
// - Individual: Each colleague has their own 24-hour timeline
// - Shared: Single timeline showing all colleagues shifted by offset

// renderTimeline renders the complete timeline view
func (m Model) renderTimeline() string {
	var b strings.Builder

	// Header
	localTime := time.Now().In(m.localTimezone)
	header := fmt.Sprintf("🌍 Timeline View - Local Time: %s (%s)",
		FormatTime(localTime, m.config.TimeFormat),
		FormatDate(localTime))
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Calculate visible range
	start := m.scrollOffset
	end := min(start+MaxVisible, len(m.colleagues))

	// Show scroll indicators
	topIndicator, bottomIndicator := renderScrollIndicators(m.scrollOffset, MaxVisible, len(m.colleagues))
	b.WriteString(topIndicator)

	// Render visible colleagues
	for i := start; i < end; i++ {
		if m.config.TimelineMode == "individual" {
			b.WriteString(m.renderTimelineRow(i, m.colleagues[i]))
		} else {
			b.WriteString(m.renderSharedTimelineRow(i, m.colleagues[i]))
		}
		b.WriteString("\n")
	}

	// Show bottom scroll indicator
	b.WriteString(bottomIndicator)

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

// renderIndividualBar generates a timeline bar for individual mode
func (m Model) renderIndividualBar(ct ColleagueTime, barWidth int) string {
	bar := make([]rune, barWidth)

	// Get current hour position (0-24)
	currentHour := float64(ct.CurrentTime.Hour())
	currentMinute := float64(ct.CurrentTime.Minute())
	currentPosition := (currentHour + currentMinute/60.0) / 24.0 // 0.0 to 1.0
	markerIndex := int(currentPosition * float64(barWidth))

	// Get colleague's hours (using accessor methods for defaults)
	workStart := ct.Colleague.GetWorkStart()
	workEnd := ct.Colleague.GetWorkEnd()
	sleepStart := ct.Colleague.GetSleepStart()
	sleepEnd := ct.Colleague.GetSleepEnd()

	// Build bar character by character
	for i := range barWidth {
		// Calculate which hour(s) this position represents
		hourFraction := float64(i) / float64(barWidth)
		hour := int(hourFraction * 24.0)

		// Determine character based on time range (marker position will be colored differently)
		if isInTimeRange(hour, sleepStart, sleepEnd) {
			bar[i] = '░' // Sleep
		} else if !ct.IsWeekend && isInTimeRange(hour, workStart, workEnd) {
			bar[i] = '█' // Work hours
		} else {
			bar[i] = '▓' // Awake off-hours
		}
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
		startPos := centerPos - (labelLen / 2) + 1

		// Ensure we don't go out of bounds
		if startPos < 1 {
			startPos = 1
		}
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
		"c cycle colors",
		"? help",
		"q quit",
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

// isInTimeRange checks if an hour falls within a time range (handles wraparound)
func isInTimeRange(hour, start, end int) bool {
	if start <= end {
		// Normal range (e.g., 9-17)
		return hour >= start && hour < end
	}
	// Wraparound range (e.g., 23-7)
	return hour >= start || hour < end
}

// truncateOrPad truncates or pads a string to exact width
func truncateOrPad(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// calculateOffsetHours calculates the hour offset between two times
func calculateOffsetHours(t time.Time, localTz *time.Location) float64 {
	localTime := time.Now().In(localTz)
	_, localOffset := localTime.Zone()
	_, remoteOffset := t.Zone()

	offsetSeconds := remoteOffset - localOffset
	return float64(offsetSeconds) / 3600.0
}

// calculateShiftAmount returns the number of bar positions to shift
// based on timezone offset. Positive offset shifts right (future),
// negative shifts left (past).
//
// Example: offset +9h with barWidth 48 returns 18 positions
func calculateShiftAmount(offsetHours float64, barWidth int) int {
	shiftFraction := offsetHours / 24.0
	shiftAmount := int(math.Round(shiftFraction * float64(barWidth)))
	return shiftAmount
}

// renderSharedTimelineHeader renders the header row for shared timeline mode
func (m Model) renderSharedTimelineHeader() string {
	barWidth := m.calculateTimelineBarWidth()

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
	// Using HH:MM format for shared mode
	hours := []int{0, 6, 12, 18, 24}
	labels := []string{"00:00", "06:00", "12:00", "18:00", "00:00"}

	for i, hour := range hours {
		// Calculate exact position for this hour in the bar
		centerPos := int(float64(hour) / 24.0 * float64(barWidth))
		label := labels[i]
		labelLen := len(label)

		// Calculate start position to center the label
		// +1 accounts for the opening bracket
		startPos := centerPos - (labelLen / 2) + 1

		// Ensure we don't go out of bounds
		if startPos < 1 {
			startPos = 1
		}
		if startPos+labelLen > barWidth+1 {
			startPos = barWidth + 1 - labelLen
		}

		// Place each character of the label
		for j, ch := range label {
			pos := startPos + j
			if pos > 0 && pos <= barWidth {
				labelChars[pos] = ch
			}
		}
	}

	// Calculate current time marker position
	localTime := time.Now().In(m.localTimezone)
	currentHour := float64(localTime.Hour())
	currentMinute := float64(localTime.Minute())
	currentPosition := (currentHour + currentMinute/60.0) / 24.0
	markerIndex := int(currentPosition * float64(barWidth))

	// Build the display
	scheme := getCurrentColorScheme(m.config.ColorScheme)

	// Left padding to align with colleague rows
	leftPadding := NameFieldWidth + 2
	padding := strings.Repeat(" ", leftPadding)

	labelLine := padding + string(labelChars)

	// Build marker label (just "now", the bars themselves show the position)
	markerPadding := strings.Repeat(" ", leftPadding+markerIndex+1)
	nowLabel := lipgloss.NewStyle().Foreground(scheme.MarkerColor).Bold(true).Render("now")
	markerLine := markerPadding + nowLabel + "\n"

	return footerStyle.Render("Local Time:") + "      " + footerStyle.Render(labelLine) + "\n" + markerLine
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

	// Calculate current time marker position (local time)
	localTime := time.Now().In(m.localTimezone)
	currentHour := float64(localTime.Hour())
	currentMinute := float64(localTime.Minute())
	currentPosition := (currentHour + currentMinute/60.0) / 24.0
	markerIndex := int(currentPosition * float64(barWidth))

	// Get colleague's hours (using accessor methods for defaults)
	workStart := ct.Colleague.GetWorkStart()
	workEnd := ct.Colleague.GetWorkEnd()
	sleepStart := ct.Colleague.GetSleepStart()
	sleepEnd := ct.Colleague.GetSleepEnd()

	// Build bar with shift applied
	for i := range barWidth {
		// Calculate which hour this position represents in local time
		localHourFraction := float64(i) / float64(barWidth)

		// Convert to their timezone hour by adding the offset
		// If they're +3h ahead, when local is 12:00, they're at 15:00
		theirHourFraction := localHourFraction + (offsetHours / 24.0)

		// Wrap around if needed
		for theirHourFraction < 0 {
			theirHourFraction += 1.0
		}
		for theirHourFraction >= 1.0 {
			theirHourFraction -= 1.0
		}

		theirHour := int(theirHourFraction * 24.0)

		// Determine character based on their local hour
		if isInTimeRange(theirHour, sleepStart, sleepEnd) {
			bar[i] = '░' // Sleep
		} else if !ct.IsWeekend && isInTimeRange(theirHour, workStart, workEnd) {
			bar[i] = '█' // Work
		} else {
			bar[i] = '▓' // Awake off
		}
	}

	// Pass markerIndex to colorizer for proper styling (will highlight that position)
	return m.colorizeBar(bar, ct, markerIndex)
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

	// Format as hours (handle half hours too)
	if offsetHours == float64(int(offsetHours)) {
		return fmt.Sprintf("%s%dh", sign, int(offsetHours))
	}
	return fmt.Sprintf("%s%.1fh", sign, offsetHours)
}
