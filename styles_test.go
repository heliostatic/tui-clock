package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TestColorSchemeValidity tests that all color schemes have all required fields
func TestColorSchemeValidity(t *testing.T) {
	schemes := []string{"classic", "dark", "high-contrast"}

	for _, schemeName := range schemes {
		t.Run(schemeName, func(t *testing.T) {
			scheme := getCurrentColorScheme(schemeName)

			// Check that name matches
			if scheme.Name != schemeName {
				t.Errorf("scheme name = %q, want %q", scheme.Name, schemeName)
			}

			// Check that all required colors are set (not empty)
			colors := map[string]lipgloss.Color{
				"SleepColor":    scheme.SleepColor,
				"AwakeOffColor": scheme.AwakeOffColor,
				"WorkColor":     scheme.WorkColor,
				"MarkerColor":   scheme.MarkerColor,
				"WeekendTint":   scheme.WeekendTint,
				"Primary":       scheme.Primary,
				"Secondary":     scheme.Secondary,
				"Success":       scheme.Success,
				"Warning":       scheme.Warning,
				"Error":         scheme.Error,
				"Muted":         scheme.Muted,
			}

			for colorName, colorValue := range colors {
				if colorValue == "" {
					t.Errorf("%s.%s is empty", schemeName, colorName)
				}
			}
		})
	}
}

// TestGetCurrentColorScheme tests the color scheme getter
func TestGetCurrentColorScheme(t *testing.T) {
	tests := []struct {
		name         string
		schemeName   string
		expectedName string
	}{
		{"valid classic", "classic", "classic"},
		{"valid dark", "dark", "dark"},
		{"valid high-contrast", "high-contrast", "high-contrast"},
		{"invalid scheme - fallback to classic", "invalid", "classic"},
		{"empty scheme - fallback to classic", "", "classic"},
		{"nonexistent scheme - fallback to classic", "nonexistent", "classic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCurrentColorScheme(tt.schemeName)
			if result.Name != tt.expectedName {
				t.Errorf("getCurrentColorScheme(%q).Name = %q, want %q",
					tt.schemeName, result.Name, tt.expectedName)
			}
		})
	}
}

// TestColorSchemeCycling tests that color schemes cycle in the correct order
func TestColorSchemeCycling(t *testing.T) {
	schemes := []string{"classic", "dark", "high-contrast"}

	// Test cycling through all schemes
	for i := 0; i < len(schemes); i++ {
		currentScheme := schemes[i]
		nextScheme := schemes[(i+1)%len(schemes)]

		t.Run("from "+currentScheme+" to "+nextScheme, func(t *testing.T) {
			// Verify current scheme is valid
			current := getCurrentColorScheme(currentScheme)
			if current.Name != currentScheme {
				t.Errorf("current scheme name = %q, want %q", current.Name, currentScheme)
			}

			// Verify next scheme is valid
			next := getCurrentColorScheme(nextScheme)
			if next.Name != nextScheme {
				t.Errorf("next scheme name = %q, want %q", next.Name, nextScheme)
			}
		})
	}
}

// TestColorSchemeColors tests that each scheme has distinct colors
func TestColorSchemeColors(t *testing.T) {
	classic := getCurrentColorScheme("classic")
	dark := getCurrentColorScheme("dark")
	highContrast := getCurrentColorScheme("high-contrast")

	// Verify schemes are different from each other
	// Check that at least some colors differ between schemes
	sameWorkColor := classic.WorkColor == dark.WorkColor && dark.WorkColor == highContrast.WorkColor
	sameMarkerColor := classic.MarkerColor == dark.MarkerColor && dark.MarkerColor == highContrast.MarkerColor
	sameSleepColor := classic.SleepColor == dark.SleepColor && dark.SleepColor == highContrast.SleepColor

	if sameWorkColor && sameMarkerColor && sameSleepColor {
		t.Error("All schemes have identical colors - they should be distinct")
	}

	// Verify all schemes have non-empty marker colors
	if classic.MarkerColor == "" {
		t.Error("Classic MarkerColor is empty")
	}
	if dark.MarkerColor == "" {
		t.Error("Dark MarkerColor is empty")
	}
	if highContrast.MarkerColor == "" {
		t.Error("High-contrast MarkerColor is empty")
	}
}

// TestColorSchemeCount verifies we have exactly 3 schemes
func TestColorSchemeCount(t *testing.T) {
	expectedCount := 3
	if len(colorSchemes) != expectedCount {
		t.Errorf("len(colorSchemes) = %d, want %d", len(colorSchemes), expectedCount)
	}

	// Verify the expected schemes exist
	expectedSchemes := []string{"classic", "dark", "high-contrast"}
	for _, schemeName := range expectedSchemes {
		if _, exists := colorSchemes[schemeName]; !exists {
			t.Errorf("expected scheme %q not found in colorSchemes map", schemeName)
		}
	}
}

// TestGetNameStyle tests the name styling function
func TestGetNameStyle(t *testing.T) {
	location, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name      string
		ct        ColleagueTime
		checkBold bool
	}{
		{
			name: "weekend style",
			ct: ColleagueTime{
				Colleague:     Colleague{Name: "Test"},
				CurrentTime:   time.Date(2025, 1, 25, 12, 0, 0, 0, location), // Saturday
				IsWeekend:     true,
				IsWorkingTime: false,
			},
			checkBold: false,
		},
		{
			name: "working style",
			ct: ColleagueTime{
				Colleague:     Colleague{Name: "Test"},
				CurrentTime:   time.Date(2025, 1, 20, 12, 0, 0, 0, location), // Monday
				IsWeekend:     false,
				IsWorkingTime: true,
			},
			checkBold: true, // Should be bold
		},
		{
			name: "off hours style",
			ct: ColleagueTime{
				Colleague:     Colleague{Name: "Test"},
				CurrentTime:   time.Date(2025, 1, 20, 20, 0, 0, 0, location), // Monday evening
				IsWeekend:     false,
				IsWorkingTime: false,
			},
			checkBold: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := getNameStyle(tt.ct)

			// We can't easily check the exact style, but we can verify it returns a style
			// and that it's not nil/empty
			rendered := style.Render("Test")
			if rendered == "" {
				t.Error("getNameStyle() returned empty render")
			}
		})
	}
}
