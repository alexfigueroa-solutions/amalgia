package main

import (
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Constants for application states
const (
	stateSelectingFiles = "selecting_files"
	stateMainMenu       = "main_menu"
	statePerforming     = "performing"
	stateSelectREADMEs  = "selecting_readmes" // New state
	stateViewingLogs    = "viewing_logs"      // New state
)

// Constants for actions
const (
	actionGenerateResume      = "generate_resume"
	actionGenerateCoverLetter = "generate_cover_letter"
	actionFetchREADMEs        = "fetch_readmes"
)

// Model represents the state of the application
type model struct {
	choices         []fs.DirEntry // Directory entries for file selection
	cursor          int
	selected        []string          // Selected files
	directory       string            // Current directory
	readmes         map[string]string // Map of README contents
	readmeList      []string          // List of README names
	selectedREADMEs map[string]bool   // Map to track selected READMEs
	state           string            // Current application state
	err             error             // Error message
	spinner         spinner.Model     // Spinner model
	spinnerActive   bool              // Spinner active status
	message         string            // Message to display
	action          string            // Current action
	startTime       time.Time         // Action start time
	// Inside the model struct, add the following fields:
	logs         []string // Slice to hold recent log messages
	logLimit     int      // Maximum number of log messages to keep
	fetchedCount int      // Number of fetched READMEs
	totalRepos   int
	failedCount  int
}

func (m model) Init() tea.Cmd {
	return nil
}

// Add the following method to the model struct
func (m *model) addLog(msg string) {
	if len(m.logs) >= m.logLimit {
		m.logs = m.logs[1:]
	}
	m.logs = append(m.logs, msg)
	logger.Println(msg) // Also write to logger
}

// Initialize the model
func initialModel() model {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Inside the initialModel function, initialize the log fields:
	return model{
		choices:         files,
		directory:       cwd,
		readmes:         make(map[string]string),
		readmeList:      []string{},
		selectedREADMEs: make(map[string]bool),
		state:           stateSelectingFiles,
		spinner:         sp,
		logs:            []string{},
		logLimit:        100, // Adjust as needed
	}

}

// Helper functions

// contains checks if a slice contains an item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// remove removes an item from a slice
func remove(slice []string, item string) []string {
	newSlice := []string{}
	for _, s := range slice {
		if s != item {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}
