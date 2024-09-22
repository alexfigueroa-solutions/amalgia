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
		s.WriteString(m.viewFileSelection())
	case stateMainMenu:
		s.WriteString(m.viewMainMenu())
	case statePerforming:
		s.WriteString(m.viewPerforming())
	case stateSelectREADMEs:
		s.WriteString(m.viewReadmeSelection())
	case stateViewingLogs:
		s.WriteString(m.viewLogs())
	}

	if m.err != nil && m.state != stateViewingLogs {
		s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return s.String()
}

// viewFileSelection renders the file selection screen
func (m *model) viewFileSelection() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("Current Directory: %s\n", m.directory)))
	s.WriteString(normalStyle.Render("Use arrow keys to navigate, space to select files, enter to confirm selection, backspace to go up a directory\n\n"))

	for i, file := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = selectedStyle.Render("❯ ")
		}
		selected := "[ ]"
		if contains(m.selected, filepath.Join(m.directory, file.Name())) {
			selected = "[x]"
		}
		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, selected, file.Name()))
	}

	if m.message != "" {
		s.WriteString("\n" + messageStyle.Render(m.message))
	}

	s.WriteString("\n\nPress 'l' to view logs.")

	return s.String()
}

// viewMainMenu renders the main menu screen
func (m *model) viewMainMenu() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("\nAI-Powered Actions:\n\n"))
	menuOptions := []string{"Generate Resume", "Generate Cover Letter", "Fetch GitHub READMEs", "View Logs", "Quit"}
	for i, option := range menuOptions {
		prefix := "  "
		if m.cursor == i {
			prefix = selectedStyle.Render("❯ ")
			s.WriteString(prefix + selectedStyle.Render(option) + "\n")
		} else {
			s.WriteString(prefix + normalStyle.Render(option) + "\n")
		}
	}

	if m.message != "" {
		s.WriteString("\n" + messageStyle.Render(m.message))
	}

	return s.String()
}

// viewPerforming renders the performing action screen
func (m *model) viewPerforming() string {
	var s strings.Builder

	if m.progressActive {
		s.WriteString("\n" + m.progress.View())
	} else if m.spinnerActive {
		s.WriteString("\n" + m.spinner.View() + " " + messageStyle.Render(m.message))
	} else {
		s.WriteString("\n" + messageStyle.Render(m.message))
	}

	return s.String()
}

// viewReadmeSelection renders the README selection screen
func (m *model) viewReadmeSelection() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Select the READMEs to include in your resume/cover letter:\n"))
	s.WriteString(normalStyle.Render("Use arrow keys to navigate, space to select/deselect, enter to confirm selection.\n\n"))

	for i, name := range m.readmeList {
		cursor := "  "
		if m.cursor == i {
			cursor = selectedStyle.Render("❯ ")
		}
		selected := "[ ]"
		if m.selectedREADMEs[name] {
			selected = "[x]"
		}
		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, selected, name))
	}

	if m.message != "" {
		s.WriteString("\n" + messageStyle.Render(m.message))
	}

	s.WriteString("\n\nPress 'l' to view logs.")

	return s.String()
}

// viewLogs renders the logs screen
func (m *model) viewLogs() string {
	var s strings.Builder

	s.WriteString(logTitleStyle.Render("=== Application Logs ===\n\n"))
	// Display the last N logs, wrapped to terminal width
	for _, logMsg := range m.logs {
		wrapped := wordwrap.String(logStyle.Render(logMsg), 80) // Adjust width as needed
		s.WriteString(wrapped + "\n")
	}
	s.WriteString("\nPress 'b' to go back to the main menu.")

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
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "space":
				file := m.choices[m.cursor]
				filePath := filepath.Join(m.directory, file.Name())
				if contains(m.selected, filePath) {
					m.selected = remove(m.selected, filePath)
					m.message = fmt.Sprintf("Deselected: %s", file.Name())
				} else {
					m.selected = append(m.selected, filePath)
					m.message = fmt.Sprintf("Selected: %s", file.Name())
				}
			case "enter":
				m.state = stateMainMenu
				m.cursor = 0
				m.message = ""
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
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < 4 { // Adjust if more menu options are added
					m.cursor++
				}
			case "enter":
				switch m.cursor {
				case 0: // Generate Resume
					m.action = actionGenerateResume
					m.state = statePerforming
					m.spinnerActive = true
					m.message = "Generating resume using OpenAI..."
					m.startTime = time.Now()
					m.addLog("Initiated resume generation.")
					return m, tea.Batch(m.spinner.Tick, generateResume(m))
				case 1: // Generate Cover Letter
					m.action = actionGenerateCoverLetter
					m.state = statePerforming
					m.spinnerActive = true
					m.message = "Generating cover letter using OpenAI..."
					m.startTime = time.Now()
					m.addLog("Initiated cover letter generation.")
					return m, tea.Batch(m.spinner.Tick, m.generateCoverLetter)
				case 2: // Fetch GitHub READMEs
					m.action = actionFetchREADMEs
					m.state = statePerforming
					m.spinnerActive = true
					m.progressActive = true
					m.message = "Fetching README files from GitHub..."
					m.startTime = time.Now()
					m.addLog("Initiated fetching GitHub READMEs.")
					return m, tea.Batch(m.spinner.Tick, m.fetchGitHubREADMEs())
				case 3: // View Logs
					m.state = stateViewingLogs
					m.cursor = 0
					m.message = ""
					m.addLog("Opened log view from main menu.")
				case 4: // Quit
					m.addLog("Application terminated by user.")
					return m, tea.Quit
				}
			case "ctrl+c", "q":
				m.addLog("Application terminated by user.")
				return m, tea.Quit
			}
		}

	case statePerforming:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmds []tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
			if m.progressActive {
				cmds = append(cmds, m.updateProgressBar())
			}
			return m, tea.Batch(cmds...)

		case FetchProgressMsg:
			// Update the progress bar with each fetched README
			m.fetchedCount++
			m.addLog(fmt.Sprintf("Progress Update: %d/%d", m.fetchedCount, m.totalRepos))
			return m, m.updateProgressBar()

		case FetchCompleteMsg:
			m.addLog("Received FetchCompleteMsg")
			m.spinnerActive = false
			m.progressActive = false
			m.state = stateSelectREADMEs
			m.cursor = 0
			m.message = "README fetching complete."
			return m, nil

		case string:
			duration := time.Since(m.startTime)
			m.spinnerActive = false
			m.progressActive = false
			m.message = fmt.Sprintf("%s\nOperation took: %v", msg, duration)
			m.state = stateMainMenu
			m.addLog(fmt.Sprintf("Completed action '%s' in %v.", m.action, duration))
			return m, nil

		case error:
			m.spinnerActive = false
			m.progressActive = false
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
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
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

// updateProgressBar updates the progress bar based on the current fetched state
func (m *model) updateProgressBar() tea.Cmd {
	if m.totalRepos == 0 {
		return nil
	}
	percent := float64(m.fetchedCount) / float64(m.totalRepos)
	m.addLog(fmt.Sprintf("Setting progress bar to %.2f%%", percent*100))
	return m.progress.SetPercent(percent)
}

// Define a new message type for fetch progress
type FetchProgressMsg struct{}
