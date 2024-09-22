package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check for required environment variables
	requiredEnvVars := []string{"OPENAI_API_KEY", "GITHUB_TOKEN"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Error: %s environment variable is not set", envVar)
		}
	}

	// Initialize and run the Bubble Tea program
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
