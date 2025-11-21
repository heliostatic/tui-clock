package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const maxVisible = 8 // Maximum colleagues visible at once

// View renders the UI
func (m Model) View() string {
	if m.inputMode == ModeHelp {
		return m.renderHelp()
	}

	var b strings.Builder

	// Header
	localTime := time.Now().In(m.localTimezone)
	header := fmt.Sprintf("🌍 World Clock - Local Time: %s (%s)",
		FormatTime(localTime, m.config.TimeFormat),
		FormatDate(localTime))
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Input mode prompts
	switch m.inputMode {
	case ModeAddName:
		b.WriteString(promptStyle.Render("Add colleague - Name: "))
		b.WriteString(m.nameInput.View())
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("Press Enter to continue, Esc to cancel"))

	case ModeSearchTimezone:
		b.WriteString(promptStyle.Render(fmt.Sprintf("Add colleague '%s' - Search timezone: ", m.nameInput.Value())))
		b.WriteString(m.searchQuery)
		b.WriteString("\n\n")
		b.WriteString(m.renderSearchResults())
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("Type to search • ↑/↓ navigate • Enter select • Esc cancel"))

	case ModeEditName:
		b.WriteString(promptStyle.Render("Edit colleague - Name: "))
		b.WriteString(m.nameInput.View())
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("Press Enter to continue, Esc to cancel"))

	case ModeEditSearchTimezone:
		b.WriteString(promptStyle.Render(fmt.Sprintf("Edit '%s' - Search timezone: ", m.nameInput.Value())))
		b.WriteString(m.searchQuery)
		b.WriteString("\n\n")
		b.WriteString(m.renderSearchResults())
		b.WriteString("\n")
		b.WriteString(footerStyle.Render("Type to search • ↑/↓ navigate • Enter select • Esc cancel"))

	default:
		// Normal mode - show colleagues
		b.WriteString(m.renderColleagues())
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.errorMsg))
	}

	// Footer with keybindings (only in normal mode)
	if m.inputMode == ModeNormal {
		b.WriteString("\n")
		b.WriteString(m.renderFooter())
	}

	return b.String()
}

// renderColleagues renders the list of colleagues with scrolling
func (m Model) renderColleagues() string {
	if len(m.colleagues) == 0 {
		return footerStyle.Render("No colleagues configured. Press 'a' to add one.")
	}

	var b strings.Builder

	// Calculate visible range
	start := m.scrollOffset
	end := start + maxVisible
	if end > len(m.colleagues) {
		end = len(m.colleagues)
	}

	// Show scroll indicator at top if needed
	if m.scrollOffset > 0 {
		b.WriteString(footerStyle.Render(fmt.Sprintf("  ▲ %d more above\n", m.scrollOffset)))
	}

	// Render visible colleagues
	for i := start; i < end; i++ {
		colleague := m.colleagues[i]
		b.WriteString(m.renderColleagueRow(i, colleague))
		b.WriteString("\n")
	}

	// Show scroll indicator at bottom if needed
	if end < len(m.colleagues) {
		remaining := len(m.colleagues) - end
		b.WriteString(footerStyle.Render(fmt.Sprintf("  ▼ %d more below\n", remaining)))
	}

	return b.String()
}

// renderColleagueRow renders a single colleague row
func (m Model) renderColleagueRow(index int, ct ColleagueTime) string {
	cursor := "  "
	style := rowStyle

	if index == m.cursor {
		cursor = "▶ "
		style = selectedRowStyle
	}

	// Status indicator (working/off-hours/weekend)
	var statusIndicator string
	var timeStyle lipgloss.Style

	if ct.IsWeekend {
		statusIndicator = "◆"
		timeStyle = weekendStyle
	} else if ct.IsWorkingTime {
		statusIndicator = "●"
		timeStyle = workingStyle
	} else {
		statusIndicator = "○"
		timeStyle = offHoursStyle
	}

	// Format time
	timeStr := FormatTime(ct.CurrentTime, m.config.TimeFormat)
	dateStr := FormatDate(ct.CurrentTime)

	// Build row
	line := fmt.Sprintf("%s%s %s  %s  %s  %s",
		cursor,
		statusIndicator,
		ct.Colleague.Name,
		timeStyle.Render(timeStr),
		offsetStyle.Render(ct.Offset),
		dateStyle.Render(dateStr),
	)

	return style.Render(line)
}

// renderSearchResults renders the timezone search results
func (m Model) renderSearchResults() string {
	if len(m.searchResults) == 0 {
		if m.searchQuery == "" {
			return footerStyle.Render("Type a city name, abbreviation (e.g., CST), or country to search...")
		}
		return footerStyle.Render(fmt.Sprintf("No results found for '%s'", m.searchQuery))
	}

	var b strings.Builder

	// Calculate visible range
	maxSearchVisible := 10
	start := m.searchScrollOffset
	end := start + maxSearchVisible
	if end > len(m.searchResults) {
		end = len(m.searchResults)
	}

	// Show scroll indicator at top if needed
	if m.searchScrollOffset > 0 {
		b.WriteString(footerStyle.Render(fmt.Sprintf("  ▲ %d more above\n", m.searchScrollOffset)))
	}

	// Render visible search results
	for i := start; i < end; i++ {
		result := m.searchResults[i]
		b.WriteString(m.renderSearchResult(i, result))
		b.WriteString("\n")
	}

	// Show scroll indicator at bottom if needed
	if end < len(m.searchResults) {
		remaining := len(m.searchResults) - end
		b.WriteString(footerStyle.Render(fmt.Sprintf("  ▼ %d more below", remaining)))
	}

	return b.String()
}

// renderSearchResult renders a single search result
func (m Model) renderSearchResult(index int, result SearchResult) string {
	cursor := "  "
	style := rowStyle

	if index == m.searchCursor {
		cursor = "▶ "
		style = selectedRowStyle
	}

	// Format: City, Country (Timezone) [Abbrevs] - Current Time
	abbrevs := ""
	if len(result.City.Abbrevs) > 0 {
		abbrevs = " [" + strings.Join(result.City.Abbrevs, "/") + "]"
	}

	timeStr := FormatTime(result.CurrentTime, m.config.TimeFormat)

	line := fmt.Sprintf("%s%s, %s (%s)%s - %s",
		cursor,
		result.City.City,
		result.City.Country,
		result.City.Timezone,
		abbrevs,
		workingStyle.Render(timeStr),
	)

	return style.Render(line)
}

// renderFooter renders the keybindings help
func (m Model) renderFooter() string {
	help := []string{
		"↑/k up",
		"↓/j down",
		"a add",
		"e edit",
		"d delete",
		"f format",
		"? help",
		"q quit",
	}
	return footerStyle.Render(strings.Join(help, " • "))
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	help := `
🌍 World Clock - Help

NAVIGATION
  ↑, k         Move cursor up
  ↓, j         Move cursor down

ACTIONS
  a            Add a new colleague
  e            Edit selected colleague
  d            Delete selected colleague
  f            Toggle time format (12h/24h)

STATUS INDICATORS
  ● Green      Working hours (9am-5pm, weekdays)
  ○ Gray       Off hours
  ◆ Purple     Weekend

GENERAL
  ?            Show this help
  q, Esc       Quit application
  Ctrl+C       Force quit

Press any key to return...
`
	return helpStyle.Render(help)
}
