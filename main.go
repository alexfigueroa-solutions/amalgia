// Filename: main.go
package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Global logger
var logger *log.Logger

func main() {
	// Initialize the logger
	logger = InitializeLogger()
	logger.Println("Application started.")

	// Check for required environment variables
	requiredEnvVars := []string{"OPENAI_API_KEY", "GITHUB_TOKEN"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			logger.Fatalf("Error: %s environment variable is not set", envVar)
		}
	}

	// Initialize and run the Bubble Tea program
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		logger.Fatalf("Error running program: %v", err)
	}
}
