package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/google/go-github/v45/github"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
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
	choices        []os.FileInfo    // List of files and directories
	cursor         int              // Current cursor position
	selected       []string         // Selected files
	directory      string           // Current directory path
	readmes        map[string]string // Map of repository names to README contents
	state          string           // Current state
	err            error            // Error message
	spinner        spinner.Model    // Spinner for indicating loading
	spinnerActive  bool             // Flag to indicate if spinner is active
	message        string           // Success or error message
	action         string           // Current action being performed
	generatedFiles []string         // List of generated files
}

// Initialize the model
func initialModel() model {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(cwd)
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

// Init initializes the Bubble Tea program
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case stateSelectingFiles:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "space":
				// Toggle selection
				selectedFile := m.choices[m.cursor]
				selectedPath := filepath.Join(m.directory, selectedFile.Name())
				if contains(m.selected, selectedPath) {
					// Deselect
					m.selected = remove(m.selected, selectedPath)
					m.message = fmt.Sprintf("Deselected: %s", selectedFile.Name())
				} else {
					// Select
					if len(m.selected) < 2 {
						m.selected = append(m.selected, selectedPath)
						m.message = fmt.Sprintf("Selected: %s", selectedFile.Name())
					} else {
						m.message = "You can select up to 2 files."
					}
				}
			case "enter":
				// Proceed to the next page even if no files are selected
				m.state = stateMainMenu
				m.message = "Proceeding without selecting files."
			case "backspace":
				// Navigate up a directory
				parentDir := filepath.Dir(m.directory)
				if parentDir != m.directory {
					newFiles, err := ioutil.ReadDir(parentDir)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.choices = newFiles
					m.directory = parentDir
					m.cursor = 0
					m.message = ""
				}
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}

	case stateMainMenu:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "1":
				// Generate Resume
				m.action = actionGenerateResume
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Generating resume using OpenAI..."
				return m, tea.Batch(m.spinner.Tick, m.generateResume)
			case "2":
				// Generate Cover Letter
				m.action = actionGenerateCoverLetter
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Generating cover letter using OpenAI..."
				return m, tea.Batch(m.spinner.Tick, m.generateCoverLetter)
			case "3":
				// Fetch GitHub READMEs
				m.action = actionFetchREADMEs
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Fetching README files from GitHub..."
				return m, tea.Batch(m.spinner.Tick, m.fetchGitHubREADMEs)
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}

	case statePerforming:
		// Update spinner
		if m.spinnerActive {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

		// Handle messages from actions
		switch msg := msg.(type) {
		case string:
			// Success or error message
			m.spinnerActive = false
			m.message = msg
			m.state = stateMainMenu
			return m, nil
		case error:
			m.spinnerActive = false
			m.err = msg
			m.state = stateMainMenu
			return m, nil
		}
	}

	// Update spinner regardless of state
	if m.spinnerActive {
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the UI based on the current state
func (m model) View() string {
	var s string

	switch m.state {
	case stateSelectingFiles:
		s += fmt.Sprintf("Current Directory: %s\n", m.directory)
		s += "Use arrow keys to navigate, space to select files, enter to confirm selection, backspace to go up a directory\n\n"
		for i, file := range m.choices {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor pointing at this choice
			}
			selected := " "
			for _, sel := range m.selected {
				if sel == filepath.Join(m.directory, file.Name()) {
					selected = "[x]"
					break
				}
			}
			s += fmt.Sprintf("%s %s %s\n", cursor, selected, file.Name())
		}
	case stateMainMenu:
		s += "Files Imported:\n"
		if len(m.selected) == 0 {
			s += "- None\n"
		} else {
			for _, file := range m.selected {
				s += fmt.Sprintf("- %s\n", file)
			}
		}
		s += "\nAI-Powered Actions:\n"
		s += "1. Generate Resume\n"
		s += "2. Generate Cover Letter\n"
		s += "3. Fetch GitHub READMEs\n"
		s += "\nPress the number of the action you want to perform, or q to quit.\n"
	default:
		// Performing state
		if m.spinnerActive {
			s += fmt.Sprintf("\n%s %s", m.spinner.View(), m.message)
		} else if m.message != "" {
			s += fmt.Sprintf("\n\n%s", m.message)
		}
	}

	// Display error message if any
	if m.err != nil {
		s += fmt.Sprintf("\n\nError: %v\n", m.err)
	}

	return s
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to remove a string from a slice
func remove(slice []string, item string) []string {
	newSlice := []string{}
	for _, s := range slice {
		if s != item {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// generateResume uses OpenAI's API to generate a resume
func (m model) generateResume() tea.Msg {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Sprintf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Combine your resume, cover letter, and README contents
	inputData, err := prepareInputData(m)
	if err != nil {
		return err
	}

	// Create a prompt for OpenAI
	prompt := fmt.Sprintf("Using the following data, generate a professional resume:\n\n%s", inputData)

	// Call OpenAI API
	resp, err := client.CreateCompletion(ctx, openai.CompletionRequest{
		Model:     openai.GPT3TextDavinci003,
		Prompt:    prompt,
		MaxTokens: 1000,
	})
	if err != nil {
		return err
	}

	// Save the generated resume
	err = ioutil.WriteFile("generated_resume.txt", []byte(resp.Choices[0].Text), 0644)
	if err != nil {
		return err
	}

	return "Resume generated successfully and saved to 'generated_resume.txt'"
}

// generateCoverLetter uses OpenAI's API to generate a cover letter
func (m model) generateCoverLetter() tea.Msg {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Sprintf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Combine your resume, cover letter, and README contents
	inputData, err := prepareInputData(m)
	if err != nil {
		return err
	}

	// Create a prompt for OpenAI
	prompt := fmt.Sprintf("Using the following data, generate a professional cover letter:\n\n%s", inputData)

	// Call OpenAI API
	resp, err := client.CreateCompletion(ctx, openai.CompletionRequest{
		Model:     openai.GPT3TextDavinci003,
		Prompt:    prompt,
		MaxTokens: 1000,
	})
	if err != nil {
		return err
	}

	// Save the generated cover letter
	err = ioutil.WriteFile("generated_cover_letter.txt", []byte(resp.Choices[0].Text), 0644)
	if err != nil {
		return err
	}

	return "Cover letter generated successfully and saved to 'generated_cover_letter.txt'"
}

// fetchGitHubREADMEs fetches README files from your GitHub repositories
func (m model) fetchGitHubREADMEs() tea.Msg {
	readmes, err := fetchGitHubREADMEs()
	if err != nil {
		return err
	}
	m.readmes = readmes
	return "GitHub README files fetched successfully and saved to 'readmes' directory"
}

// fetchGitHubREADMEs fetches README files from your GitHub repositories
func fetchGitHubREADMEs() (map[string]string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Get the authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	// Prepare the directory to save READMEs
	readmesDir := "readmes"
	if err := os.MkdirAll(readmesDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory '%s': %v", readmesDir, err)
	}

	// List repositories
	repos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Visibility:  "all",
		Affiliation: "owner",
	})
	if err != nil {
		return nil, err
	}

	readmeContents := make(map[string]string)

	for _, repo := range repos {
		// Get the README
		readme, resp, err := client.Repositories.GetReadme(ctx, *user.Login, *repo.Name, nil)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				// Skip repositories without a README
				continue
			}
			log.Printf("Error fetching README for %s: %v", *repo.Name, err)
			continue
		}

		content, err := readme.GetContent()
		if err != nil {
			log.Printf("Error decoding README for %s: %v", *repo.Name, err)
			continue
		}

		// Save the README content to a file
		filename := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", *repo.Name))
		err = ioutil.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			log.Printf("Error writing README for %s to file: %v", *repo.Name, err)
			continue
		}

		// Store the content in the map
		readmeContents[*repo.Name] = content
	}

	return readmeContents, nil
}

// prepareInputData combines selected files and README contents
func prepareInputData(m model) (string, error) {
	var buffer bytes.Buffer

	// Add selected file contents (if any)
	if len(m.selected) == 0 {
		buffer.WriteString("No resume or cover letter provided.\n\n")
	} else {
		for _, file := range m.selected {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				return "", err
			}
			buffer.WriteString(fmt.Sprintf("File: %s\n", filepath.Base(file)))
			buffer.Write(content)
			buffer.WriteString("\n\n")
		}
	}

	// Add README contents (if any)
	if len(m.readmes) == 0 {
		buffer.WriteString("No GitHub README files found.\n\n")
	} else {
		for repoName, content := range m.readmes {
			buffer.WriteString(fmt.Sprintf("Project: %s\n", repoName))
			buffer.WriteString(content)
			buffer.WriteString("\n\n")
		}
	}

	return buffer.String(), nil
}

// main function starts the Bubble Tea program
func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
