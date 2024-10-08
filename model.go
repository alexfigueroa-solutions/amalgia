package main

import (
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Constants for application states
const (
	stateSelectingFiles  = "selecting_files"
	stateMainMenu        = "main_menu"
	statePerforming      = "performing"
	stateSelectREADMEs   = "selecting_readmes"
	stateViewingLogs     = "viewing_logs"
	stateChatWithProfile = "chat_with_profile" // New state
)

// Constants for actions
const (
	actionGenerateResume      = "generate_resume"
	actionGenerateCoverLetter = "generate_cover_letter"
	actionFetchREADMEs        = "fetch_readmes"
	actionChatWithProfile     = "chat_with_profile" // New action
)

// Styles using lipgloss
var (
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00CED1"))
	menuStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	selectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	normalStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	logTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA07A"))
	logStyle         = lipgloss.NewStyle().PaddingLeft(2)
	messageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	statusBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#FFFFFF"))
	spinnerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	progressBarStyle = progress.WithScaledGradient("#FF7F50", "#FF6347")
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
	progress        progress.Model    // Progress bar model
	progressActive  bool              // Progress bar active status
	message         string            // Message to display
	action          string            // Current action
	startTime       time.Time         // Action start time
	logs            []string          // Slice to hold recent log messages
	logLimit        int               // Maximum number of log messages to keep
	fetchedCount    int               // Number of fetched READMEs
	totalRepos      int
	failedCount     int
	program         *tea.Program
	chatHistory     []string // Stores the chat messages
	chatInput       string   // Stores the current user input
}

// Init is the first method that gets called. It sets up the model.
func (m *model) Init() tea.Cmd {
	// Return any initial commands to run
	return nil
}

// Initialize the model
func initialModel(p *tea.Program) *model {
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
	sp.Spinner = spinner.Line
	sp.Style = spinnerStyle

	// Initialize progress bar
	pr := progress.New(progressBarStyle)

	return &model{
		choices:         files,
		directory:       cwd,
		readmes:         make(map[string]string),
		readmeList:      []string{},
		selectedREADMEs: make(map[string]bool),
		state:           stateSelectingFiles,
		spinner:         sp,
		progress:        pr,
		logs:            []string{},
		logLimit:        100, // Adjust as needed
		program:         p,
		chatHistory:     []string{},
		chatInput:       "",
	}
}

func (m *model) addLog(msg string) {
	if len(m.logs) >= m.logLimit {
		m.logs = m.logs[1:]
	}
	m.logs = append(m.logs, msg)
	logger.Println(msg) // Also write to logger
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
