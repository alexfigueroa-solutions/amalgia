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
)

// Constants for actions
const (
	actionGenerateResume      = "generate_resume"
	actionGenerateCoverLetter = "generate_cover_letter"
	actionFetchREADMEs        = "fetch_readmes"
)

// Model represents the state of the application
type model struct {
	choices       []fs.DirEntry // Changed from []os.FileInfo to []fs.DirEntry
	cursor        int
	selected      []string
	directory     string
	readmes       map[string]string
	state         string
	err           error
	spinner       spinner.Model
	spinnerActive bool
	message       string
	action        string
	startTime     time.Time
}

func (m model) Init() tea.Cmd {
	return nil
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

	return model{
		choices:   files,
		directory: cwd,
		readmes:   make(map[string]string),
		state:     stateSelectingFiles,
		spinner:   sp,
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
