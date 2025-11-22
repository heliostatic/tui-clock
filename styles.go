package main

import (
	"sort"

	"github.com/charmbracelet/lipgloss"
)

// ColorScheme defines a complete set of colors for the application
// Colors can be either lipgloss.Color (simple) or lipgloss.AdaptiveColor (light/dark aware)
type ColorScheme struct {
	Name string

	// Timeline colors - use TerminalColor to support both regular and adaptive colors
	SleepColor    lipgloss.TerminalColor
	AwakeOffColor lipgloss.TerminalColor
	WorkColor     lipgloss.TerminalColor
	MarkerColor   lipgloss.TerminalColor
	WeekendTint   lipgloss.TerminalColor

	// UI colors
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Success   lipgloss.TerminalColor
	Warning   lipgloss.TerminalColor
	Error     lipgloss.TerminalColor
	Muted     lipgloss.TerminalColor
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
	// Classic - Vibrant colors with true color support
	classicScheme = ColorScheme{
		Name:          "classic",
		SleepColor:    lipgloss.Color("#585858"), // Dark gray
		AwakeOffColor: lipgloss.Color("#a8a8a8"), // Medium gray
		WorkColor:     lipgloss.Color("#00d787"), // Bright green
		MarkerColor:   lipgloss.Color("#00d7ff"), // Bright cyan
		WeekendTint:   lipgloss.Color("#af87d7"), // Purple
		Primary:       lipgloss.Color("#00d7ff"), // Bright cyan
		Secondary:     lipgloss.Color("#ff87d7"), // Pink
		Success:       lipgloss.Color("#00d787"), // Bright green
		Warning:       lipgloss.Color("#ffaf00"), // Orange
		Error:         lipgloss.Color("#ff0000"), // Red
		Muted:         lipgloss.Color("#585858"), // Gray
	}

	// Dark - Muted night-mode colors with true color
	darkScheme = ColorScheme{
		Name:          "dark",
		SleepColor:    lipgloss.Color("#1c1c1c"), // Very dark
		AwakeOffColor: lipgloss.Color("#444444"), // Dark gray
		WorkColor:     lipgloss.Color("#5f875f"), // Muted green
		MarkerColor:   lipgloss.Color("#5f87af"), // Muted cyan
		WeekendTint:   lipgloss.Color("#875f87"), // Muted purple
		Primary:       lipgloss.Color("#5f87af"), // Muted cyan
		Secondary:     lipgloss.Color("#af5f87"), // Muted pink
		Success:       lipgloss.Color("#5f875f"), // Muted green
		Warning:       lipgloss.Color("#af875f"), // Muted orange
		Error:         lipgloss.Color("#af5f5f"), // Muted red
		Muted:         lipgloss.Color("#3a3a3a"), // Dark gray
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

	// Nord - Nordic theme with adaptive light/dark support
	// Uses true colors from https://www.nordtheme.com/docs/colors-and-palettes
	nordScheme = ColorScheme{
		Name: "nord",
		// Adaptive: darker sleep in dark mode, lighter in light mode
		SleepColor: lipgloss.AdaptiveColor{
			Light: "#e5e9f0", // Snow Storm nord5
			Dark:  "#2e3440", // Polar Night nord0 (darkest)
		},
		AwakeOffColor: lipgloss.AdaptiveColor{
			Light: "#d8dee9", // Snow Storm nord4
			Dark:  "#4c566a", // Polar Night nord3
		},
		WorkColor:     lipgloss.Color("#a3be8c"), // Aurora green nord14
		MarkerColor:   lipgloss.Color("#88c0d0"), // Frost cyan nord8
		WeekendTint:   lipgloss.Color("#b48ead"), // Aurora purple nord15
		Primary:       lipgloss.Color("#88c0d0"), // Frost cyan nord8
		Secondary:     lipgloss.Color("#81a1c1"), // Frost blue nord9
		Success:       lipgloss.Color("#a3be8c"), // Aurora green nord14
		Warning:       lipgloss.Color("#ebcb8b"), // Aurora yellow nord13
		Error:         lipgloss.Color("#bf616a"), // Aurora red nord11
		Muted:         lipgloss.Color("#4c566a"), // Polar Night nord3
	}

	// Solarized - Precision colors for reduced eye strain
	// Adaptive scheme using official Solarized colors
	// https://github.com/altercation/solarized#the-values
	solarizedScheme = ColorScheme{
		Name: "solarized",
		// Base3/Base03 for sleep (lightest/darkest backgrounds)
		SleepColor: lipgloss.AdaptiveColor{
			Light: "#fdf6e3", // Base3 (light background)
			Dark:  "#002b36", // Base03 (dark background)
		},
		// Base2/Base02 for off-hours
		AwakeOffColor: lipgloss.AdaptiveColor{
			Light: "#eee8d5", // Base2
			Dark:  "#073642", // Base02
		},
		WorkColor:   lipgloss.Color("#859900"), // Green
		MarkerColor: lipgloss.Color("#cb4b16"), // Orange
		WeekendTint: lipgloss.Color("#d33682"), // Magenta
		Primary:     lipgloss.Color("#268bd2"), // Blue
		Secondary:   lipgloss.Color("#2aa198"), // Cyan
		Success:     lipgloss.Color("#859900"), // Green
		Warning:     lipgloss.Color("#b58900"), // Yellow
		Error:       lipgloss.Color("#dc322f"), // Red
		// Base0/Base00 for muted text
		Muted: lipgloss.AdaptiveColor{
			Light: "#657b83", // Base00
			Dark:  "#839496", // Base0
		},
	}

	colorSchemes = map[string]ColorScheme{
		"classic":       classicScheme,
		"dark":          darkScheme,
		"high-contrast": highContrastScheme,
		"nord":          nordScheme,
		"solarized":     solarizedScheme,
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
	if scheme.SleepColor == nil {
		missing = append(missing, "SleepColor")
	}
	if scheme.AwakeOffColor == nil {
		missing = append(missing, "AwakeOffColor")
	}
	if scheme.WorkColor == nil {
		missing = append(missing, "WorkColor")
	}
	if scheme.MarkerColor == nil {
		missing = append(missing, "MarkerColor")
	}
	if scheme.WeekendTint == nil {
		missing = append(missing, "WeekendTint")
	}
	if scheme.Primary == nil {
		missing = append(missing, "Primary")
	}
	if scheme.Secondary == nil {
		missing = append(missing, "Secondary")
	}
	if scheme.Success == nil {
		missing = append(missing, "Success")
	}
	if scheme.Warning == nil {
		missing = append(missing, "Warning")
	}
	if scheme.Error == nil {
		missing = append(missing, "Error")
	}
	if scheme.Muted == nil {
		missing = append(missing, "Muted")
	}

	return missing
}
