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

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	choices       []os.FileInfo
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
				selectedFile := m.choices[m.cursor]
				selectedPath := filepath.Join(m.directory, selectedFile.Name())
				if contains(m.selected, selectedPath) {
					m.selected = remove(m.selected, selectedPath)
					m.message = fmt.Sprintf("Deselected: %s", selectedFile.Name())
				} else {
					if len(m.selected) < 2 {
						m.selected = append(m.selected, selectedPath)
						m.message = fmt.Sprintf("Selected: %s", selectedFile.Name())
					} else {
						m.message = "You can select up to 2 files."
					}
				}
			case "enter":
				m.state = stateMainMenu
				m.message = "Proceeding to main menu."
			case "backspace":
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
				m.action = actionGenerateResume
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Generating resume using OpenAI..."
				m.startTime = time.Now()
				return m, tea.Batch(m.spinner.Tick, m.generateResume)
			case "2":
				m.action = actionGenerateCoverLetter
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Generating cover letter using OpenAI..."
				m.startTime = time.Now()
				return m, tea.Batch(m.spinner.Tick, m.generateCoverLetter)
			case "3":
				m.action = actionFetchREADMEs
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Fetching README files from GitHub..."
				m.startTime = time.Now()
				return m, tea.Batch(m.spinner.Tick, m.fetchGitHubREADMEs)
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}

	case statePerforming:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case string:
			duration := time.Since(m.startTime)
			m.spinnerActive = false
			m.message = fmt.Sprintf("%s\nOperation took: %v", msg, duration)
			m.state = stateMainMenu
			return m, nil
		case error:
			m.spinnerActive = false
			m.err = msg
			m.message = fmt.Sprintf("Error: %v", msg)
			m.state = stateMainMenu
			return m, nil
		}
	}

	return m, cmd
}

// View renders the UI based on the current state
func (m model) View() string {
	var s string

	switch m.state {
	case stateSelectingFiles:
		s += fmt.Sprintf("Current Directory: %s\n", m.directory)
		s += "Use arrow keys to navigate, space to select files, enter to confirm selection, backspace to go up a directory\n\n"
		for i, file := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			selected := " "
			if contains(m.selected, filepath.Join(m.directory, file.Name())) {
				selected = "[x]"
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
	case statePerforming:
		if m.spinnerActive {
			s += fmt.Sprintf("\n%s %s", m.spinner.View(), m.message)
		} else {
			s += fmt.Sprintf("\n%s", m.message)
		}
	}

	if m.err != nil {
		s += fmt.Sprintf("\n\nError: %v\n", m.err)
	}

	return s
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	newSlice := []string{}
	for _, s := range slice {
		if s != item {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// generateResume uses OpenAI's Chat Completions API (GPT-4) to generate a resume
func (m model) generateResume() tea.Msg {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	inputData, err := prepareInputData(m)
	if err != nil {
		return fmt.Errorf("error preparing input data: %v", err)
	}

	req := openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a professional resume writer. You may not have all the information but dissect the project readmes and generate a professional resume anyways. I work at company_name for 3 years now as an associate software develpoer doing full stack work btw. You should heavily include my projects in the resume.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Using the following data, generate a professional resume:\n\n%s", inputData),
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("error generating resume: %v", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from GPT-4")
	}

	err = ioutil.WriteFile("generated_resume.txt", []byte(resp.Choices[0].Message.Content), 0644)
	if err != nil {
		return fmt.Errorf("error saving resume: %v", err)
	}

	return "Resume generated and saved to 'generated_resume.txt'"
}

// generateCoverLetter uses OpenAI's API to generate a cover letter
// generateCoverLetter uses OpenAI's API to generate a cover letter
func (m model) generateCoverLetter() tea.Msg {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	inputData, err := prepareInputData(m)
	if err != nil {
		return fmt.Errorf("error preparing input data: %v", err)
	}

	req := openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a professional cover letter writer. Generate a compelling cover letter based on the provided information. Tailor the letter to highlight the candidate's skills and experiences that are most relevant to a software development position.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Using the following data, generate a professional cover letter:\n\n%s", inputData),
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("error generating cover letter: %v", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from GPT-4")
	}

	err = ioutil.WriteFile("generated_cover_letter.txt", []byte(resp.Choices[0].Message.Content), 0644)
	if err != nil {
		return fmt.Errorf("error saving cover letter: %v", err)
	}

	return "Cover letter generated and saved to 'generated_cover_letter.txt'"
}

// fetchGitHubREADMEs fetches README files from your GitHub repositories
func (m *model) fetchGitHubREADMEs() tea.Msg {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	readmesDir := "readmes"
	if err := os.MkdirAll(readmesDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory '%s': %v", readmesDir, err)
	}

	repos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Visibility:  "all",
		Affiliation: "owner",
	})
	if err != nil {
		return fmt.Errorf("error listing repositories: %v", err)
	}

	readmeContents := make(map[string]string)

	for _, repo := range repos {
		readme, resp, err := client.Repositories.GetReadme(ctx, *user.Login, *repo.Name, nil)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
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

		filename := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", *repo.Name))
		err = ioutil.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			log.Printf("Error writing README for %s to file: %v", *repo.Name, err)
			continue
		}

		readmeContents[*repo.Name] = content
	}

	m.readmes = readmeContents
	return fmt.Sprintf("Fetched and saved README files for %d repositories", len(readmeContents))
}

// prepareInputData combines selected files and README contents
func prepareInputData(m model) (string, error) {
	var buffer bytes.Buffer

	if len(m.selected) == 0 {
		buffer.WriteString("No resume or cover letter provided.\n\n")
	} else {
		for _, file := range m.selected {
			content, err := ioutil.ReadFile(file)
			if err != nil {
				return "", fmt.Errorf("error reading file %s: %v", file, err)
			}
			buffer.WriteString(fmt.Sprintf("File: %s\n", filepath.Base(file)))
			buffer.Write(content)
			buffer.WriteString("\n\n")
		}
	}

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
