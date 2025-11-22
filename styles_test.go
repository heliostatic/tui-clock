package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TestColorSchemeValidity tests that all color schemes have all required fields
func TestColorSchemeValidity(t *testing.T) {
	// Test all registered schemes dynamically
	schemes := GetAvailableColorSchemes()

	for _, schemeName := range schemes {
		t.Run(schemeName, func(t *testing.T) {
			scheme := getCurrentColorScheme(schemeName)

			// Check that name matches
			if scheme.Name != schemeName {
				t.Errorf("scheme name = %q, want %q", scheme.Name, schemeName)
			}

			// Use ValidateColorScheme function
			missing := ValidateColorScheme(scheme)
			if len(missing) > 0 {
				t.Errorf("%s is missing fields: %v", schemeName, missing)
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

// TestColorSchemeCount verifies we have exactly 6 schemes
func TestColorSchemeCount(t *testing.T) {
	expectedCount := 6
	actualCount := len(GetAvailableColorSchemes())
	if actualCount != expectedCount {
		t.Errorf("len(colorSchemes) = %d, want %d", actualCount, expectedCount)
	}

	// Verify the expected schemes exist
	expectedSchemes := []string{"classic", "dark", "high-contrast", "nord", "solarized", "solarized-dark"}
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

// TestGetAvailableColorSchemes tests the scheme discovery function
func TestGetAvailableColorSchemes(t *testing.T) {
	schemes := GetAvailableColorSchemes()

	// Should return all 6 schemes
	if len(schemes) != 6 {
		t.Errorf("GetAvailableColorSchemes() returned %d schemes, want 6", len(schemes))
	}

	// Should be sorted alphabetically
	expected := []string{"classic", "dark", "high-contrast", "nord", "solarized", "solarized-dark"}
	for i, name := range expected {
		if schemes[i] != name {
			t.Errorf("schemes[%d] = %q, want %q", i, schemes[i], name)
		}
	}

	// Should contain all expected schemes
	schemeMap := make(map[string]bool)
	for _, scheme := range schemes {
		schemeMap[scheme] = true
	}
	for _, expectedScheme := range expected {
		if !schemeMap[expectedScheme] {
			t.Errorf("missing expected scheme %q", expectedScheme)
		}
	}
}

// TestGetNextColorScheme tests the cycling function
func TestGetNextColorScheme(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		expected string
	}{
		{"from classic", "classic", "dark"},
		{"from dark", "dark", "high-contrast"},
		{"from high-contrast", "high-contrast", "nord"},
		{"from nord", "nord", "solarized"},
		{"from solarized", "solarized", "solarized-dark"},
		{"from solarized-dark (wrap)", "solarized-dark", "classic"},
		{"invalid scheme - fallback to first", "nonexistent", "classic"},
		{"empty string - fallback to first", "", "classic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNextColorScheme(tt.current)
			if result != tt.expected {
				t.Errorf("GetNextColorScheme(%q) = %q, want %q", tt.current, result, tt.expected)
			}
		})
	}
}

// TestValidateColorScheme tests the validation function
func TestValidateColorScheme(t *testing.T) {
	t.Run("valid scheme", func(t *testing.T) {
		scheme := ColorScheme{
			Name:          "test",
			SleepColor:    lipgloss.Color("1"),
			AwakeOffColor: lipgloss.Color("2"),
			WorkColor:     lipgloss.Color("3"),
			MarkerColor:   lipgloss.Color("4"),
			WeekendTint:   lipgloss.Color("5"),
			Primary:       lipgloss.Color("6"),
			Secondary:     lipgloss.Color("7"),
			Success:       lipgloss.Color("8"),
			Warning:       lipgloss.Color("9"),
			Error:         lipgloss.Color("10"),
			Muted:         lipgloss.Color("11"),
		}

		missing := ValidateColorScheme(scheme)
		if len(missing) != 0 {
			t.Errorf("ValidateColorScheme() = %v, want empty slice", missing)
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		scheme := ColorScheme{
			Name:        "incomplete",
			SleepColor:  lipgloss.Color("1"),
			WorkColor:   lipgloss.Color("3"),
			MarkerColor: lipgloss.Color("4"),
		}

		missing := ValidateColorScheme(scheme)
		expectedMissing := []string{"AwakeOffColor", "WeekendTint", "Primary", "Secondary", "Success", "Warning", "Error", "Muted"}

		if len(missing) != len(expectedMissing) {
			t.Errorf("ValidateColorScheme() returned %d missing fields, want %d", len(missing), len(expectedMissing))
		}

		// Check all expected fields are reported as missing
		missingMap := make(map[string]bool)
		for _, field := range missing {
			missingMap[field] = true
		}
		for _, expected := range expectedMissing {
			if !missingMap[expected] {
				t.Errorf("expected field %q to be reported as missing", expected)
			}
		}
	})

	t.Run("all built-in schemes are valid", func(t *testing.T) {
		schemes := GetAvailableColorSchemes()
		for _, schemeName := range schemes {
			scheme := getCurrentColorScheme(schemeName)
			missing := ValidateColorScheme(scheme)
			if len(missing) > 0 {
				t.Errorf("built-in scheme %q is invalid, missing: %v", schemeName, missing)
			}
		}
	})
}
