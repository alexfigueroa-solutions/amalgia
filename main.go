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

	// Create a new Bubble Tea program and pass it to the model
	var p *tea.Program
	m := initialModel(p)
	p = tea.NewProgram(m)
	m.program = p // Now set the program in the model

	// Run the Bubble Tea program
	if _, err := p.Run(); err != nil {
		logger.Fatalf("Error running program: %v", err)
	}
}
