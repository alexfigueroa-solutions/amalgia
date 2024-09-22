package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"
)

func generateResume(m *model) tea.Cmd {
	return func() tea.Msg {
		m.addLog("Starting resume generation.")
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			errMsg := "OPENAI_API_KEY environment variable not set"
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		client := openai.NewClient(apiKey)
		ctx := context.Background()

		inputData, err := prepareInputData(m)
		if err != nil {
			errMsg := fmt.Sprintf("Error preparing input data: %v", err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		req := openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a professional resume writer. You will not have all the context you need, but do the best you can use the context of the readmes and project to extrapolate and write good detailed prject sections. Make sure its structured like a resume and only shows the most prominent projects. Extrapolate all the other sections based on the info you have. Make sure to include the most relevant projects and skills.",
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
			errMsg := fmt.Sprintf("Error generating resume: %v", err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		if len(resp.Choices) == 0 {
			errMsg := "No response from GPT-4"
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		err = os.WriteFile("generated_resume.txt", []byte(resp.Choices[0].Message.Content), 0600)
		if err != nil {
			errMsg := fmt.Sprintf("Error saving resume: %v", err)
			m.addLog(errMsg)
			return fmt.Errorf(errMsg)
		}

		successMsg := "Resume generated and saved to 'generated_resume.txt'"
		m.addLog(successMsg)
		return successMsg
	}
}

func (m *model) generateCoverLetter() tea.Msg {
	m.addLog("Starting cover letter generation.")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		errMsg := "OPENAI_API_KEY environment variable not set"
		m.addLog(errMsg)
		return fmt.Errorf(errMsg)
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	inputData, err := prepareInputData(m)
	if err != nil {
		errMsg := fmt.Sprintf("Error preparing input data: %v", err)
		m.addLog(errMsg)
		return fmt.Errorf(errMsg)
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
		errMsg := fmt.Sprintf("Error generating cover letter: %v", err)
		m.addLog(errMsg)
		return fmt.Errorf(errMsg)
	}

	if len(resp.Choices) == 0 {
		errMsg := "No response from GPT-4"
		m.addLog(errMsg)
		return fmt.Errorf(errMsg)
	}

	err = os.WriteFile("generated_cover_letter.txt", []byte(resp.Choices[0].Message.Content), 0600)
	if err != nil {
		errMsg := fmt.Sprintf("Error saving cover letter: %v", err)
		m.addLog(errMsg)
		return fmt.Errorf(errMsg)
	}

	successMsg := "Cover letter generated and saved to 'generated_cover_letter.txt'"
	m.addLog(successMsg)
	return successMsg
}

func prepareInputData(m *model) (string, error) {
	var buffer bytes.Buffer

	if len(m.selected) == 0 {
		buffer.WriteString("No additional files provided.\n\n")
	} else {
		for _, file := range m.selected {
			content, err := os.ReadFile(file)
			if err != nil {
				errMsg := fmt.Sprintf("Error reading file %s: %v", file, err)
				m.addLog(errMsg)
				return "", fmt.Errorf(errMsg)
			}
			buffer.WriteString(fmt.Sprintf("File: %s\n", filepath.Base(file)))
			buffer.Write(content)
			buffer.WriteString("\n\n")
		}
	}

	// Include selected READMEs
	selectedReadmeCount := 0
	for name, selected := range m.selectedREADMEs {
		if selected {
			selectedReadmeCount++
			content, ok := m.readmes[name]
			if !ok {
				errMsg := fmt.Sprintf("README content for %s not found", name)
				m.addLog(errMsg)
				return "", fmt.Errorf(errMsg)
			}
			buffer.WriteString(fmt.Sprintf("Project: %s\n", name))
			buffer.WriteString(content)
			buffer.WriteString("\n\n")
		}
	}

	if selectedReadmeCount == 0 {
		buffer.WriteString("No GitHub README files selected.\n\n")
	}

	return buffer.String(), nil
}
