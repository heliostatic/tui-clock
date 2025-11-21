package main

import (
	"strings"
	"testing"
)

func TestSearchTimezones(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectAtLeast  int
		expectContains []string // City names that should be in results
		expectFirst    string   // What should be the top result
	}{
		{
			name:           "Empty query returns all",
			query:          "",
			expectAtLeast:  100, // Should return many cities
			expectContains: []string{},
			expectFirst:    "", // Don't check first for empty query
		},
		{
			name:           "Search 'london'",
			query:          "london",
			expectAtLeast:  1,
			expectContains: []string{"London"},
			expectFirst:    "London",
		},
		{
			name:           "Search 'new york'",
			query:          "new york",
			expectAtLeast:  1,
			expectContains: []string{"New York"},
			expectFirst:    "New York",
		},
		{
			name:           "Search 'tokyo'",
			query:          "tokyo",
			expectAtLeast:  1,
			expectContains: []string{"Tokyo"},
			expectFirst:    "Tokyo",
		},
		{
			name:           "Search 'cst' - ambiguous abbreviation",
			query:          "cst",
			expectAtLeast:  3, // Should find Chicago, Shanghai, Havana
			expectContains: []string{"Chicago", "Shanghai", "Havana"},
			expectFirst:    "", // Order may vary, don't check
		},
		{
			name:           "Search 'est'",
			query:          "est",
			expectAtLeast:  5,
			expectContains: []string{"New York", "Miami", "Toronto"},
			expectFirst:    "", // Many EST cities
		},
		{
			name:           "Search 'japan'",
			query:          "japan",
			expectAtLeast:  1,
			expectContains: []string{"Tokyo", "Osaka"},
			expectFirst:    "", // Multiple Japan cities
		},
		{
			name:           "Search 'australia'",
			query:          "australia",
			expectAtLeast:  3,
			expectContains: []string{"Sydney", "Melbourne", "Brisbane"},
			expectFirst:    "Sydney", // Most populous
		},
		{
			name:           "Search 'lincoln'",
			query:          "lincoln",
			expectAtLeast:  1,
			expectContains: []string{"Lincoln"},
			expectFirst:    "Lincoln",
		},
		{
			name:          "Invalid search",
			query:         "xyznonexistent",
			expectAtLeast: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := SearchTimezones(tt.query)

			if len(results) < tt.expectAtLeast {
				t.Errorf("Expected at least %d results, got %d", tt.expectAtLeast, len(results))
			}

			// Check that expected cities are in results
			for _, expectedCity := range tt.expectContains {
				found := false
				for _, result := range results {
					if result.City.City == expectedCity {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find '%s' in results, but didn't", expectedCity)
				}
			}

			// Check first result if specified
			if tt.expectFirst != "" && len(results) > 0 {
				if results[0].City.City != tt.expectFirst {
					t.Errorf("Expected first result to be '%s', got '%s'", tt.expectFirst, results[0].City.City)
				}
			}
		})
	}
}

func TestSearchTimezonesCaseInsensitive(t *testing.T) {
	queries := []string{"london", "LONDON", "LoNdOn", "London"}

	var firstResults []SearchResult
	for i, query := range queries {
		results := SearchTimezones(query)
		if len(results) == 0 {
			t.Fatalf("Query '%s' returned no results", query)
		}
		if i == 0 {
			firstResults = results
		} else {
			// All queries should return same results
			if len(results) != len(firstResults) {
				t.Errorf("Case sensitivity issue: '%s' returned %d results, expected %d",
					query, len(results), len(firstResults))
			}
		}
	}
}

func TestGetDisplayNameForColleague(t *testing.T) {
	testCity := CityTimezone{
		City:     "New York",
		Country:  "United States",
		Timezone: "America/New_York",
		Abbrevs:  []string{"EST", "EDT"},
	}

	tests := []struct {
		name          string
		baseName      string
		searchQuery   string
		displayFormat string
		expected      string
	}{
		{
			name:          "Auto mode - city search",
			baseName:      "Alice",
			searchQuery:   "new york",
			displayFormat: "auto",
			expected:      "Alice (New York)",
		},
		{
			name:          "Auto mode - abbreviation search",
			baseName:      "Bob",
			searchQuery:   "est",
			displayFormat: "auto",
			expected:      "Bob (EST)",
		},
		{
			name:          "City format explicit",
			baseName:      "Charlie",
			searchQuery:   "est",
			displayFormat: "city",
			expected:      "Charlie (New York)",
		},
		{
			name:          "Timezone format explicit",
			baseName:      "Diana",
			searchQuery:   "new york",
			displayFormat: "timezone",
			expected:      "Diana (America/New_York)",
		},
		{
			name:          "Abbreviation format explicit",
			baseName:      "Eve",
			searchQuery:   "new york",
			displayFormat: "abbreviation",
			expected:      "Eve (EST)",
		},
		{
			name:          "Name already contains city",
			baseName:      "Alice (New York)",
			searchQuery:   "new york",
			displayFormat: "auto",
			expected:      "Alice (New York)", // Should not duplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDisplayNameForColleague(tt.baseName, testCity, tt.searchQuery, tt.displayFormat)
			if result != tt.expected {
				t.Errorf("GetDisplayNameForColleague() = '%s', want '%s'", result, tt.expected)
			}
		})
	}
}

func TestSearchResultsRanking(t *testing.T) {
	// Test that "york" prioritizes "New York" over "York, UK"
	results := SearchTimezones("york")

	if len(results) < 1 {
		t.Fatal("Expected at least one result for 'york'")
	}

	// First result should be New York (higher popularity)
	if results[0].City.City != "New York" {
		t.Errorf("Expected 'New York' to rank higher than other York cities, got '%s'",
			results[0].City.City)
	}
}

func TestSearchByCountry(t *testing.T) {
	results := SearchTimezones("germany")

	if len(results) == 0 {
		t.Fatal("Expected results for 'germany'")
	}

	// All results should be in Germany
	for _, result := range results {
		if !strings.Contains(strings.ToLower(result.City.Country), "german") {
			t.Errorf("Expected all results to be in Germany, found %s, %s",
				result.City.City, result.City.Country)
		}
	}
}
