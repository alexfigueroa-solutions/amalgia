package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"
)

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
				Content: "You are a professional resume writer. You may not have all the information but dissect the project readmes and generate a professional resume anyways. I work at company_name for 3 years now as an associate software developer doing full stack work btw. You should heavily include my projects in the resume.",
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
