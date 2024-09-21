package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/google/go-github/v45/github"
    "golang.org/x/oauth2"
)

func main() {
    p := tea.NewProgram(initialModel())

    if err := p.Start(); err != nil {
        fmt.Println("Error running program:", err)
        os.Exit(1)
    }
}

type model struct {
    choice  int
    files   []string
    readmes map[string]string // Map of repository names to README contents
    ready   bool
    err     error
}

func initialModel() model {
    return model{
        choice:  -1,
        files:   []string{},
        readmes: make(map[string]string),
        ready:   false,
        err:     nil,
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        }

    case error:
        m.err = msg
        return m, nil
    }

    if !m.ready {
        // Load files and GitHub data
        m.files = getResumeAndCoverLetter()
        readmes, err := fetchGitHubREADMEs()
        if err != nil {
            m.err = err
        } else {
            m.readmes = readmes
        }
        m.ready = true
    }

    return m, nil
}

func (m model) View() string {
    if m.err != nil {
        return fmt.Sprintf("An error occurred: %v\n", m.err)
    }

    if !m.ready {
        return "Loading...\n"
    }

    s := "Files Imported:\n"
    for _, file := range m.files {
        s += fmt.Sprintf("- %s\n", file)
    }
    s += "\nProjects Fetched from GitHub and READMEs saved:\n"
    for repoName := range m.readmes {
        s += fmt.Sprintf("- %s\n", repoName)
    }
    s += "\nPress q to quit.\n"

    return s
}

// getResumeAndCoverLetter is a placeholder function
// TODO: Implement file selection logic using Bubble Tea's filetree component
func getResumeAndCoverLetter() []string {
    // Placeholder: Implement logic to select files using file explorer
    return []string{
        "/path/to/resume.pdf",
        "/path/to/cover_letter.pdf",
    }
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
