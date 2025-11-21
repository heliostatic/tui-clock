package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("86")   // Cyan
	secondaryColor = lipgloss.Color("212")  // Pink
	successColor   = lipgloss.Color("42")   // Green
	warningColor   = lipgloss.Color("214")  // Orange
	errorColor     = lipgloss.Color("196")  // Red
	mutedColor     = lipgloss.Color("240")  // Gray
	weekendColor   = lipgloss.Color("141")  // Purple

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
