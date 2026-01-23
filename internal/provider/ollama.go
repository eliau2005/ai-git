package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaProvider struct {
	BaseURL      string
	Model        string
	SystemPrompt string
	CommitPrompt string
}

func (p *OllamaProvider) GetName() string {
	return "ollama"
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (p *OllamaProvider) GenerateCommitMessage(diff string, context string) (string, error) {
	// Truncate
	if len(diff) > 15000 {
		diff = diff[:15000] + "\n... [Diff truncated] ..."
	}

	url := p.BaseURL
	if url == "" {
		url = "http://localhost:11434/api/generate"
	} else {
		url = url + "/api/generate"
	}

	systemPrompt := p.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are an expert developer. Generate a raw git commit message. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks."
	}

	commitPromptTemplate := p.CommitPrompt
	if commitPromptTemplate == "" {
		commitPromptTemplate = "Generate a raw git commit message for the changes below. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks.\n\nChanges:\n%s\n\n%s"
	}

	prompt := fmt.Sprintf(commitPromptTemplate, diff, context)

	reqBody := ollamaGenerateRequest{
		Model:  p.Model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ollamaGenerateResponse
		body, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return "", fmt.Errorf("Ollama API error: %s", errResp.Error)
		}
		return "", fmt.Errorf("Ollama API error: %s", string(body))
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", result.Error)
	}

	return result.Response, nil
}
