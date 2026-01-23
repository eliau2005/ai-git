package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaProvider struct {
	BaseURL string
	Model   string
}

func (p *OllamaProvider) GetName() string {
	return "ollama"
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

func (p *OllamaProvider) GenerateCommitMessage(diff string, context string) (string, error) {
	url := p.BaseURL
	if url == "" {
		url = "http://localhost:11434/api/generate"
	} else {
		url = url + "/api/generate"
	}

	prompt := fmt.Sprintf("You are an expert developer. Generate a raw git commit message for the changes below. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks.\n\nChanges:\n%s\n\n%s", diff, context)

	reqBody := ollamaGenerateRequest{
		Model:  p.Model,
		Prompt: prompt,
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
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %s", string(body))
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}
