package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v45/github"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
)

type model struct {
	choices      []os.FileInfo // List of files and directories
	cursor       int           // Which item our cursor is pointing at
	selected     []string      // List of selected files
	directory    string        // Current directory path
	readmes      map[string]string
	state        string
	err          error
}

func initialModel() model {
	// Set up the file picker starting at the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	return model{
		choices:   files,
		directory: cwd,
		readmes:   make(map[string]string),
		state:     "selecting_files",
	}
}

func (m model) Init() tea.Cmd {
	// Initialize the file picker model
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Navigate up or down
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		// Select a file or directory
		case "enter":
			selectedFile := m.choices[m.cursor]
			selectedPath := filepath.Join(m.directory, selectedFile.Name())
			if selectedFile.IsDir() {
				// If it's a directory, navigate into it
				newFiles, err := ioutil.ReadDir(selectedPath)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.choices = newFiles
				m.directory = selectedPath
				m.cursor = 0 // Reset cursor when entering a new directory
			} else {
				// Select the file
				m.selected = append(m.selected, selectedPath)
				if len(m.selected) >= 2 {
					// Move to next state once we have two files selected
					m.state = "main_menu"
				}
			}
		case "backspace":
			// Go up a directory
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
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	var s string

	switch m.state {
	case "selecting_files":
		s += fmt.Sprintf("Current Directory: %s\n", m.directory)
		s += "Use arrow keys to navigate, enter to select, backspace to go up a directory\n"
		for i, file := range m.choices {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor pointing at this choice
			}
			s += fmt.Sprintf("%s %s\n", cursor, file.Name())
		}
	case "main_menu":
		s += "Files Imported:\n"
		for _, file := range m.selected {
			s += fmt.Sprintf("- %s\n", file)
		}
		s += "\nAI-Powered Actions:\n"
		s += "1. Generate Resume\n"
		s += "2. Generate Cover Letter\n"
		s += "3. Fetch GitHub READMEs\n"
		s += "\nPress the number of the action you want to perform, or q to quit."
	}

	return s
}

// fetchGitHubREADMEs fetches README files from your GitHub repositories
// and saves them to the 'readmes' directory
func fetchGitHubREADMEs() (map[string]string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	ctx := context.Background()
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
		readme, _, err := client.Repositories.GetReadme(ctx, *user.Login, *repo.Name, nil)
		if err != nil {
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

// generateResume uses OpenAI's API to generate a resume
func generateResume(m model) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
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

	fmt.Println("Generated resume saved to 'generated_resume.txt'")
	return nil
}

// prepareInputData combines resume, cover letter, and README contents
func prepareInputData(m model) (string, error) {
	var buffer bytes.Buffer

	// Add selected file contents
	for _, file := range m.selected {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		buffer.WriteString(fmt.Sprintf("File: %s\n", file))
		buffer.Write(content)
		buffer.WriteString("\n\n")
	}

	return buffer.String(), nil
}

func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
