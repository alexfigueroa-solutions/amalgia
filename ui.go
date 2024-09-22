// Filename: ui.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

// View renders the UI based on the current state
func (m *model) View() string {
	var s strings.Builder

	switch m.state {
	case stateSelectingFiles:
		s.WriteString(fmt.Sprintf("Current Directory: %s\n", m.directory))
		s.WriteString("Use arrow keys to navigate, space to select files, enter to confirm selection, backspace to go up a directory\n\n")
		for i, file := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			selected := " "
			if contains(m.selected, filepath.Join(m.directory, file.Name())) {
				selected = "[x]"
			}
			s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, selected, file.Name()))
		}
		s.WriteString("\n" + m.message)
		s.WriteString("\n\nPress 'l' to view logs.")
	case stateMainMenu:
		s.WriteString("Files Imported:\n")
		if len(m.selected) == 0 {
			s.WriteString("- None\n")
		} else {
			for _, file := range m.selected {
				s.WriteString(fmt.Sprintf("- %s\n", file))
			}
		}
		s.WriteString("\nAI-Powered Actions:\n")
		s.WriteString("1. Generate Resume\n")
		s.WriteString("2. Generate Cover Letter\n")
		s.WriteString("3. Fetch GitHub READMEs\n")
		s.WriteString("4. View Logs\n") // New Action
		s.WriteString("\nPress the number of the action you want to perform, or q to quit.\n")
	case statePerforming:
		if m.spinnerActive {
			s.WriteString(fmt.Sprintf("\n%s %s", m.spinner.View(), m.message))
		} else {
			s.WriteString(fmt.Sprintf("\n%s", m.message))
		}
	case stateSelectREADMEs:
		s.WriteString("Select the READMEs to include in your resume/cover letter:\n")
		s.WriteString("Use arrow keys to navigate, space to select/deselect, enter to confirm selection.\n\n")
		for i, name := range m.readmeList {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			selected := "[ ]"
			if m.selectedREADMEs[name] {
				selected = "[x]"
			}
			s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, selected, name))
		}
		s.WriteString("\n" + m.message)
		s.WriteString("\n\nPress 'l' to view logs.")
	case stateViewingLogs:
		s.WriteString("=== Application Logs ===\n\n")
		// Display the last N logs, wrapped to terminal width
		for _, logMsg := range m.logs {
			wrapped := wordwrap.String(logMsg, 80) // Adjust width as needed
			s.WriteString(wrapped + "\n")
		}
		s.WriteString("\nPress 'b' to go back to the main menu.")
	}

	if m.err != nil && m.state != stateViewingLogs {
		s.WriteString(fmt.Sprintf("\n\nError: %v\n", m.err))
	}

	return s.String()
}

// Update handles incoming messages and updates the model accordingly
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				name := m.readmeList[m.cursor]
				m.selectedREADMEs[name] = !m.selectedREADMEs[name]

				if m.selectedREADMEs[name] {
					m.message = fmt.Sprintf("Selected: %s", name)
					m.addLog(fmt.Sprintf("Selected README: %s", name))
				} else {
					m.message = fmt.Sprintf("Deselected: %s", name)
					m.addLog(fmt.Sprintf("Deselected README: %s", name))
				}

			case "enter":
				m.state = stateMainMenu
				m.cursor = 0
				m.message = "Proceeding to main menu."
				m.addLog("Navigated to main menu.")
			case "backspace":
				parentDir := filepath.Dir(m.directory)
				if parentDir != m.directory {
					newFiles, err := os.ReadDir(parentDir)
					if err != nil {
						m.err = err
						m.addLog(fmt.Sprintf("Error navigating to parent directory: %v", err))
						return m, nil
					}
					m.choices = newFiles
					m.directory = parentDir
					m.cursor = 0
					m.message = ""
					m.addLog(fmt.Sprintf("Navigated to parent directory: %s", parentDir))
				}
			case "l":
				m.state = stateViewingLogs
				m.cursor = 0
				m.message = ""
				m.addLog("Opened log view.")
			case "ctrl+c", "q":
				m.addLog("Application terminated by user.")
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
				m.addLog("Initiated resume generation.")
				return m, tea.Batch(m.spinner.Tick, generateResume(m))
			case "2":
				m.action = actionGenerateCoverLetter
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Generating cover letter using OpenAI..."
				m.startTime = time.Now()
				m.addLog("Initiated cover letter generation.")
				return m, tea.Batch(m.spinner.Tick, m.generateCoverLetter)
			case "3":
				m.action = actionFetchREADMEs
				m.state = statePerforming
				m.spinnerActive = true
				m.message = "Fetching README files from GitHub..."
				m.startTime = time.Now()
				m.addLog("Initiated fetching GitHub READMEs.")
				return m, tea.Batch(m.spinner.Tick, m.fetchGitHubREADMEs())
			case "4":
				m.state = stateViewingLogs
				m.cursor = 0
				m.message = ""
				m.addLog("Opened log view from main menu.")
			case "ctrl+c", "q":
				m.addLog("Application terminated by user.")
				return m, tea.Quit
			}
		}

	case statePerforming:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case FetchCompleteMsg:
			m.addLog("Received FetchCompleteMsg")
			m.spinnerActive = false
			m.state = stateSelectREADMEs
			m.cursor = 0
			m.message = "README fetching complete."
			return m, nil

		case string:
			if m.action == actionFetchREADMEs {
				// Transition to README selection
				m.spinnerActive = false
				m.state = stateSelectREADMEs
				m.cursor = 0
				m.message = ""
				m.addLog("Completed fetching GitHub READMEs.")
				return m, nil
			} else {
				duration := time.Since(m.startTime)
				m.spinnerActive = false
				m.message = fmt.Sprintf("%s\nOperation took: %v", msg, duration)
				m.state = stateMainMenu
				m.addLog(fmt.Sprintf("Completed action '%s' in %v.", m.action, duration))
				return m, nil
			}
		case error:
			m.spinnerActive = false
			m.err = msg
			m.message = fmt.Sprintf("Error: %v", msg)
			m.state = stateMainMenu
			m.addLog(fmt.Sprintf("Error during action '%s': %v", m.action, msg))
			return m, nil
		}

	case stateSelectREADMEs:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.readmeList)-1 {
					m.cursor++
				}
			case " ":
				name := m.readmeList[m.cursor]
				m.selectedREADMEs[name] = !m.selectedREADMEs[name]
				if m.selectedREADMEs[name] {
					m.message = fmt.Sprintf("Selected: %s", name)
					m.addLog(fmt.Sprintf("Selected README: %s", name))
				} else {
					m.message = fmt.Sprintf("Deselected: %s", name)
					m.addLog(fmt.Sprintf("Deselected README: %s", name))
				}
				return m, nil
			case "enter":
				m.state = stateMainMenu
				m.cursor = 0
				m.message = "Proceeding to main menu."
				m.addLog("Returned to main menu from README selection.")
				return m, nil
			case "l":
				m.state = stateViewingLogs
				m.cursor = 0
				m.message = ""
				m.addLog("Opened log view from README selection.")
				return m, nil
			case "ctrl+c", "q":
				m.addLog("Application terminated by user.")
				return m, tea.Quit
			}
		}
		return m, nil

	case stateViewingLogs:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "b":
				m.state = stateMainMenu
				m.cursor = 0
				m.message = "Returning to main menu."
				m.addLog("Closed log view and returned to main menu.")
			case "ctrl+c", "q":
				m.addLog("Application terminated by user.")
				return m, tea.Quit
			}
		}
	}

	return m, cmd
}
