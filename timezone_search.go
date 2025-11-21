package main

import (
	"sort"
	"strings"
	"time"
)

// SearchResult represents a search result with scoring
type SearchResult struct {
	City        CityTimezone
	CurrentTime time.Time
	Score       int    // Higher is better
	MatchField  string // What matched: "city", "country", "abbrev", "timezone"
}

// SearchTimezones searches the city database and returns ranked results
func SearchTimezones(query string) []SearchResult {
	if query == "" {
		// Return all cities sorted by popularity
		results := make([]SearchResult, 0, len(AllCities))
		for _, city := range AllCities {
			loc, err := time.LoadLocation(city.Timezone)
			if err != nil {
				continue
			}
			results = append(results, SearchResult{
				City:        city,
				CurrentTime: time.Now().In(loc),
				Score:       (6 - city.Popularity) * 10, // Popularity-based score
				MatchField:  "all",
			})
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		return results
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	results := make([]SearchResult, 0)

	for _, city := range AllCities {
		score := scoreMatch(city, queryLower)
		if score > 0 {
			loc, err := time.LoadLocation(city.Timezone)
			if err != nil {
				continue
			}
			matchField := getMatchField(city, queryLower)
			results = append(results, SearchResult{
				City:        city,
				CurrentTime: time.Now().In(loc),
				Score:       score,
				MatchField:  matchField,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// scoreMatch calculates a relevance score for a city against a query
func scoreMatch(city CityTimezone, queryLower string) int {
	cityLower := strings.ToLower(city.City)
	countryLower := strings.ToLower(city.Country)
	timezoneLower := strings.ToLower(city.Timezone)

	score := 0

	// Exact matches (highest priority)
	if cityLower == queryLower {
		score += 1000
	}
	if countryLower == queryLower {
		score += 800
	}
	for _, abbrev := range city.Abbrevs {
		if strings.ToLower(abbrev) == queryLower {
			score += 900
		}
	}

	// Starts with query
	if strings.HasPrefix(cityLower, queryLower) {
		score += 500
	}
	if strings.HasPrefix(countryLower, queryLower) {
		score += 400
	}

	// Contains query
	if strings.Contains(cityLower, queryLower) {
		score += 300
	}
	if strings.Contains(countryLower, queryLower) {
		score += 250
	}
	if strings.Contains(timezoneLower, queryLower) {
		score += 200
	}

	// Check abbreviations for contains
	for _, abbrev := range city.Abbrevs {
		abbrevLower := strings.ToLower(abbrev)
		if strings.Contains(abbrevLower, queryLower) {
			score += 350
		}
	}

	// Boost by popularity (1=major city gets more boost)
	if score > 0 {
		popularityBoost := (6 - city.Popularity) * 10
		score += popularityBoost
	}

	return score
}

// getMatchField determines what field matched for display purposes
func getMatchField(city CityTimezone, queryLower string) string {
	cityLower := strings.ToLower(city.City)
	countryLower := strings.ToLower(city.Country)

	if cityLower == queryLower || strings.HasPrefix(cityLower, queryLower) {
		return "city"
	}

	for _, abbrev := range city.Abbrevs {
		if strings.ToLower(abbrev) == queryLower {
			return "abbrev"
		}
	}

	if countryLower == queryLower || strings.HasPrefix(countryLower, queryLower) {
		return "country"
	}

	if strings.Contains(cityLower, queryLower) {
		return "city"
	}

	return "timezone"
}

// GetDisplayNameForColleague generates the display name with location
// Based on what the user searched for and what they selected
func GetDisplayNameForColleague(baseName string, selectedCity CityTimezone, searchQuery string, displayFormat string) string {
	queryLower := strings.ToLower(strings.TrimSpace(searchQuery))

	// If display format is explicit, use it
	switch displayFormat {
	case "city":
		return formatWithCity(baseName, selectedCity)
	case "timezone":
		return formatWithTimezone(baseName, selectedCity)
	case "abbreviation":
		return formatWithAbbrev(baseName, selectedCity)
	}

	// Auto mode: determine based on what user searched
	matchField := getMatchField(selectedCity, queryLower)

	switch matchField {
	case "abbrev":
		return formatWithAbbrev(baseName, selectedCity)
	case "city", "country":
		return formatWithCity(baseName, selectedCity)
	default:
		// Default to city if available, otherwise timezone
		return formatWithCity(baseName, selectedCity)
	}
}

func formatWithCity(baseName string, city CityTimezone) string {
	// Check if city name is already in the base name
	if strings.Contains(strings.ToLower(baseName), strings.ToLower(city.City)) {
		return baseName
	}
	return baseName + " (" + city.City + ")"
}

func formatWithTimezone(baseName string, city CityTimezone) string {
	if strings.Contains(baseName, city.Timezone) {
		return baseName
	}
	return baseName + " (" + city.Timezone + ")"
}

func formatWithAbbrev(baseName string, city CityTimezone) string {
	if len(city.Abbrevs) == 0 {
		return formatWithCity(baseName, city)
	}
	abbrev := city.Abbrevs[0] // Use first abbreviation
	if strings.Contains(strings.ToUpper(baseName), abbrev) {
		return baseName
	}
	return baseName + " (" + abbrev + ")"
}
