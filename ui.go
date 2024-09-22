package main

import (
	"fmt"
	"os" // Updated import
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Assuming you have defined your model and other necessary types elsewhere
// For example:
// type model struct {
//     state          stateType
//     directory      string
//     choices        []fs.DirEntry
//     cursor         int
//     selected       []string
//     message        string
//     err            error
//     spinner        spinner.Model
//     spinnerActive  bool
//     action         actionType
//     startTime      time.Time
// }

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
					// Use os.ReadDir instead of ioutil.ReadDir
					newFiles, err := os.ReadDir(parentDir)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.choices = newFiles // Now types match: []fs.DirEntry
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
