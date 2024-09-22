// Filename: github.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// FetchCompleteMsg is sent when all READMEs have been processed
type FetchCompleteMsg struct{}

// fetchGitHubREADMEs fetches README files from your GitHub repositories
func (m *model) fetchGitHubREADMEs() tea.Cmd {
	return func() tea.Msg {
		m.addLog("Starting to fetch GitHub READMEs.")
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			errMsg := "GITHUB_TOKEN environment variable not set"
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			errMsg := fmt.Sprintf("Error getting user: %v", err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		readmesDir := "readmes"
		if err := os.MkdirAll(readmesDir, os.ModePerm); err != nil {
			errMsg := fmt.Sprintf("Failed to create directory '%s': %v", readmesDir, err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		repos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
			Visibility:  "public",
			Affiliation: "owner",
		})
		if err != nil {
			errMsg := fmt.Sprintf("Error listing repositories: %v", err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		m.totalRepos = len(repos)
		m.fetchedCount = 0
		m.failedCount = 0

		readmeContents := make(map[string]string)
		var readmeNames []string
		var mu sync.Mutex
		var wg sync.WaitGroup

		progressChan := make(chan tea.Msg)

		// Goroutines to fetch READMEs concurrently
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
							// No README found; skip silently
							progressChan <- FetchProgressMsg{}
							return
						}
						errMsg := fmt.Sprintf("Error fetching README for %s: %v", *repo.Name, err)
						m.addLog(errMsg)
						mu.Lock()
						m.failedCount++
						mu.Unlock()
						progressChan <- FetchProgressMsg{}
						return
					}

					content, err := readme.GetContent()
					if err != nil {
						errMsg := fmt.Sprintf("Error decoding README for %s: %v", *repo.Name, err)
						m.addLog(errMsg)
						mu.Lock()
						m.failedCount++
						mu.Unlock()
						progressChan <- FetchProgressMsg{}
						return
					}

					filename := filepath.Join(readmesDir, fmt.Sprintf("%s_README.md", *repo.Name))
					err = os.WriteFile(filename, []byte(content), 0600)
					if err != nil {
						errMsg := fmt.Sprintf("Error writing README for %s to file: %v", *repo.Name, err)
						m.addLog(errMsg)
						mu.Lock()
						m.failedCount++
						mu.Unlock()
						progressChan <- FetchProgressMsg{}
						return
					}

					mu.Lock()
					readmeContents[*repo.Name] = content
					readmeNames = append(readmeNames, *repo.Name)
					m.fetchedCount++
					mu.Unlock()

					m.addLog(fmt.Sprintf("Fetched README for repository: %s", *repo.Name))
					progressChan <- FetchProgressMsg{}
				}
			}(repo)
		}

		// Wait for all fetches to complete
		go func() {
			wg.Wait()
			close(progressChan)
		}()

		// Process progress updates
		for msg := range progressChan {
			// Send message to the program
			m.program.Send(msg)
		}

		// After all goroutines are done, send a fetch complete message
		m.readmes = readmeContents
		m.readmeList = readmeNames

		return FetchCompleteMsg{}
	}
}
