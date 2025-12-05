package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ramayac/multi-cmd/internal/config"
	"github.com/ramayac/multi-cmd/internal/tui"
)

func main() {
	// Default configuration
	configPath := "commands.yaml"
	scanPath := "."
	outputPath := fmt.Sprintf("multi-cmd-results-%s.md", time.Now().Format("2006-01-02-150405"))

	// Parse command line arguments
	if len(os.Args) > 1 {
		scanPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	}
	if len(os.Args) > 3 {
		outputPath = os.Args[3]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		log.Fatalf("Invalid scan path: %v", err)
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Clear the console before starting
	fmt.Print("\033[H\033[2J")

	// Initialize and run TUI
	p := tea.NewProgram(tui.NewModel(absPath, configPath, outputPath, cfg))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
