package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timed out while fetching READMEs")
		default:
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
	}

	m.readmes = readmeContents
	return fmt.Sprintf("Fetched and saved README files for %d repositories", len(readmeContents))
}
