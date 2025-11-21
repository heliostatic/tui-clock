package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse CLI flags
	configPath := flag.String("config", "", "Path to config file (default: ~/.config/tui-clock/config.yaml)")
	flag.Parse()

	// Determine config path
	var finalConfigPath string
	if *configPath != "" {
		finalConfigPath = *configPath
	} else {
		defaultPath, err := GetDefaultConfigPath()
		if err != nil {
			fmt.Printf("Error getting default config path: %v\n", err)
			os.Exit(1)
		}
		finalConfigPath = defaultPath
	}

	// Load config
	config, err := LoadConfig(finalConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create model
	model := NewModel(config, finalConfigPath)

	// Create program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
