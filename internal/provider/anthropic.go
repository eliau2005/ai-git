package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AnthropicProvider struct {
	APIKey string
	Model  string
}

func (p *AnthropicProvider) GetName() string {
	return "anthropic"
}

type anthropicMessagesRequest struct {
	Model     string `json:"model"`
	System    string `json:"system,omitempty"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (p *AnthropicProvider) GenerateCommitMessage(diff string, context string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	systemPrompt := "You are an expert developer. Generate a raw git commit message. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks."
	userPrompt := fmt.Sprintf("Generate a git commit message for these changes:\n\n%s\n\n%s", diff, context)

	reqBody := anthropicMessagesRequest{
		Model:     p.Model,
		System:    systemPrompt,
		MaxTokens: 1024,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: userPrompt},
		},
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
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result anthropicMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return "", fmt.Errorf("no response from Anthropic")
}
