package main

import (
	"sort"

	"github.com/charmbracelet/lipgloss"
)

// ColorScheme defines a complete set of colors for the application
type ColorScheme struct {
	Name string

	// Timeline colors
	SleepColor    lipgloss.Color
	AwakeOffColor lipgloss.Color
	WorkColor     lipgloss.Color
	MarkerColor   lipgloss.Color
	WeekendTint   lipgloss.Color

	// UI colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color
}

var (
	// Color palette
	primaryColor   = lipgloss.Color("86")  // Cyan
	secondaryColor = lipgloss.Color("212") // Pink
	successColor   = lipgloss.Color("42")  // Green
	warningColor   = lipgloss.Color("214") // Orange
	errorColor     = lipgloss.Color("196") // Red
	mutedColor     = lipgloss.Color("240") // Gray
	weekendColor   = lipgloss.Color("141") // Purple

	// Header style
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// Normal colleague row
	rowStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Selected colleague row
	selectedRowStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(primaryColor).
				Bold(true)

	// Working hours indicator
	workingStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// Off-hours indicator
	offHoursStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Weekend indicator
	weekendStyle = lipgloss.NewStyle().
			Foreground(weekendColor)

	// Offset style
	offsetStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Date style
	dateStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Footer/help style
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			MarginTop(1)

	// Input prompt style
	promptStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1)
)

// Color scheme definitions
var (
	classicScheme = ColorScheme{
		Name:          "classic",
		SleepColor:    lipgloss.Color("240"), // Dark gray
		AwakeOffColor: lipgloss.Color("248"), // Medium gray
		WorkColor:     lipgloss.Color("42"),  // Green
		MarkerColor:   lipgloss.Color("86"),  // Cyan
		WeekendTint:   lipgloss.Color("141"), // Purple
		Primary:       lipgloss.Color("86"),  // Cyan
		Secondary:     lipgloss.Color("212"), // Pink
		Success:       lipgloss.Color("42"),  // Green
		Warning:       lipgloss.Color("214"), // Orange
		Error:         lipgloss.Color("196"), // Red
		Muted:         lipgloss.Color("240"), // Gray
	}

	darkScheme = ColorScheme{
		Name:          "dark",
		SleepColor:    lipgloss.Color("234"), // Very dark
		AwakeOffColor: lipgloss.Color("238"), // Dark gray
		WorkColor:     lipgloss.Color("71"),  // Muted green
		MarkerColor:   lipgloss.Color("67"),  // Muted cyan
		WeekendTint:   lipgloss.Color("96"),  // Muted purple
		Primary:       lipgloss.Color("67"),  // Muted cyan
		Secondary:     lipgloss.Color("132"), // Muted pink
		Success:       lipgloss.Color("71"),  // Muted green
		Warning:       lipgloss.Color("136"), // Muted orange
		Error:         lipgloss.Color("131"), // Muted red
		Muted:         lipgloss.Color("237"), // Dark gray
	}

	highContrastScheme = ColorScheme{
		Name:          "high-contrast",
		SleepColor:    lipgloss.Color("0"),  // Black
		AwakeOffColor: lipgloss.Color("15"), // White
		WorkColor:     lipgloss.Color("10"), // Bright green
		MarkerColor:   lipgloss.Color("11"), // Bright yellow
		WeekendTint:   lipgloss.Color("13"), // Bright magenta
		Primary:       lipgloss.Color("14"), // Bright cyan
		Secondary:     lipgloss.Color("15"), // White
		Success:       lipgloss.Color("10"), // Bright green
		Warning:       lipgloss.Color("11"), // Bright yellow
		Error:         lipgloss.Color("9"),  // Bright red
		Muted:         lipgloss.Color("8"),  // Gray
	}

	nordScheme = ColorScheme{
		Name:          "nord",
		SleepColor:    lipgloss.Color("235"), // Polar Night darkest (#2E3440)
		AwakeOffColor: lipgloss.Color("238"), // Polar Night lighter (#4C566A)
		WorkColor:     lipgloss.Color("108"), // Aurora green (#A3BE8C)
		MarkerColor:   lipgloss.Color("110"), // Frost bright cyan (#88C0D0)
		WeekendTint:   lipgloss.Color("139"), // Aurora purple (#B48EAD)
		Primary:       lipgloss.Color("110"), // Frost cyan (#88C0D0)
		Secondary:     lipgloss.Color("109"), // Frost blue (#81A1C1)
		Success:       lipgloss.Color("108"), // Aurora green (#A3BE8C)
		Warning:       lipgloss.Color("222"), // Aurora yellow (#EBCB8B)
		Error:         lipgloss.Color("167"), // Aurora red (#BF616A)
		Muted:         lipgloss.Color("243"), // Snow Storm (#D8DEE9)
	}

	solarizedScheme = ColorScheme{
		Name:          "solarized",
		SleepColor:    lipgloss.Color("254"), // Base2 (#EEE8D5)
		AwakeOffColor: lipgloss.Color("245"), // Base1 (#93A1A1)
		WorkColor:     lipgloss.Color("64"),  // Green (#859900)
		MarkerColor:   lipgloss.Color("166"), // Orange (#CB4B16)
		WeekendTint:   lipgloss.Color("125"), // Magenta (#D33682)
		Primary:       lipgloss.Color("33"),  // Blue (#268BD2)
		Secondary:     lipgloss.Color("37"),  // Cyan (#2AA198)
		Success:       lipgloss.Color("64"),  // Green (#859900)
		Warning:       lipgloss.Color("136"), // Yellow (#B58900)
		Error:         lipgloss.Color("160"), // Red (#DC322F)
		Muted:         lipgloss.Color("246"), // Base0 (#657B83)
	}

	solarizedDarkScheme = ColorScheme{
		Name:          "solarized-dark",
		SleepColor:    lipgloss.Color("234"), // Base03 (#002B36)
		AwakeOffColor: lipgloss.Color("240"), // Base01 (#586E75)
		WorkColor:     lipgloss.Color("64"),  // Green (#859900)
		MarkerColor:   lipgloss.Color("166"), // Orange (#CB4B16)
		WeekendTint:   lipgloss.Color("125"), // Magenta (#D33682)
		Primary:       lipgloss.Color("33"),  // Blue (#268BD2)
		Secondary:     lipgloss.Color("37"),  // Cyan (#2AA198)
		Success:       lipgloss.Color("64"),  // Green (#859900)
		Warning:       lipgloss.Color("136"), // Yellow (#B58900)
		Error:         lipgloss.Color("160"), // Red (#DC322F)
		Muted:         lipgloss.Color("241"), // Base00 (#657B83)
	}

	colorSchemes = map[string]ColorScheme{
		"classic":        classicScheme,
		"dark":           darkScheme,
		"high-contrast":  highContrastScheme,
		"nord":           nordScheme,
		"solarized":      solarizedScheme,
		"solarized-dark": solarizedDarkScheme,
	}
)

// getCurrentColorScheme returns the color scheme by name, or classic as fallback
func getCurrentColorScheme(schemeName string) ColorScheme {
	scheme, exists := colorSchemes[schemeName]
	if !exists {
		return classicScheme // fallback
	}
	return scheme
}

// GetAvailableColorSchemes returns all registered color scheme names, sorted alphabetically
func GetAvailableColorSchemes() []string {
	schemes := make([]string, 0, len(colorSchemes))
	for name := range colorSchemes {
		schemes = append(schemes, name)
	}
	// Sort alphabetically for predictable cycling order
	sort.Strings(schemes)
	return schemes
}

// GetNextColorScheme returns the next scheme in alphabetical order (wraps around)
func GetNextColorScheme(current string) string {
	schemes := GetAvailableColorSchemes()

	// Find current scheme's index
	for i, name := range schemes {
		if name == current {
			// Return next scheme (wrap around to first if at end)
			return schemes[(i+1)%len(schemes)]
		}
	}

	// Fallback if current scheme not found
	if len(schemes) > 0 {
		return schemes[0]
	}
	return "classic"
}

// ValidateColorScheme checks if a scheme has all required fields
// Returns a list of missing field names (empty if valid)
func ValidateColorScheme(scheme ColorScheme) []string {
	var missing []string

	if scheme.Name == "" {
		missing = append(missing, "Name")
	}
	if scheme.SleepColor == "" {
		missing = append(missing, "SleepColor")
	}
	if scheme.AwakeOffColor == "" {
		missing = append(missing, "AwakeOffColor")
	}
	if scheme.WorkColor == "" {
		missing = append(missing, "WorkColor")
	}
	if scheme.MarkerColor == "" {
		missing = append(missing, "MarkerColor")
	}
	if scheme.WeekendTint == "" {
		missing = append(missing, "WeekendTint")
	}
	if scheme.Primary == "" {
		missing = append(missing, "Primary")
	}
	if scheme.Secondary == "" {
		missing = append(missing, "Secondary")
	}
	if scheme.Success == "" {
		missing = append(missing, "Success")
	}
	if scheme.Warning == "" {
		missing = append(missing, "Warning")
	}
	if scheme.Error == "" {
		missing = append(missing, "Error")
	}
	if scheme.Muted == "" {
		missing = append(missing, "Muted")
	}

	return missing
}
