package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// fetchGitHubREADMEs fetches README files from your GitHub repositories
func (m *model) fetchGitHubREADMEs() tea.Msg {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
	var readmeNames []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(repo *github.Repository) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				readme, resp, err := client.Repositories.GetReadme(ctx, *user.Login, *repo.Name, nil)
				if err != nil {
					if resp != nil && resp.StatusCode == http.StatusNotFound {
						return
					}
					log.Printf("Error fetching README for %s: %v", *repo.Name, err)
					return
				}

				content, err := readme.GetContent()
				if err != nil {
					log.Printf("Error decoding README for %s: %v", *repo.Name, err)
					return
				}

				filename := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", *repo.Name))
				err = os.WriteFile(filename, []byte(content), 0600)
				if err != nil {
					log.Printf("Error writing README for %s to file: %v", *repo.Name, err)
					return
				}

				mu.Lock()
				readmeContents[*repo.Name] = content
				readmeNames = append(readmeNames, *repo.Name)
				mu.Unlock()
			}
		}(repo)
	}

	wg.Wait()

	if len(readmeNames) == 0 {
		m.state = stateMainMenu
		m.spinnerActive = false
		return "No READMEs found to select."
	}

	m.readmes = readmeContents
	m.readmeList = readmeNames

	// Initialize all READMEs as selected by default
	for _, name := range readmeNames {
		m.selectedREADMEs[name] = true
	}

	// Transition to README selection state
	m.state = stateSelectREADMEs
	m.cursor = 0
	m.message = ""

	return nil
}
