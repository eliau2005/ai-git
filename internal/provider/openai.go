package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {

	APIKey       string

	Model        string

	SystemPrompt string

	CommitPrompt string

}



func (p *OpenAIProvider) GetName() string {

	return "openai"

}



type openAIChatCompletionRequest struct {

	Model    string `json:"model"`

	Messages []struct {

		Role    string `json:"role"`

		Content string `json:"content"`

	} `json:"messages"`

}



type openAIChatCompletionResponse struct {

	Choices []struct {

		Message struct {

			Content string `json:"content"`

		} `json:"message"`

	} `json:"choices"`

	Error struct {

		Message string `json:"message"`

		Type    string `json:"type"`

		Code    interface{} `json:"code"` // Can be string or null

	} `json:"error,omitempty"`

}



func (p *OpenAIProvider) GenerateCommitMessage(diff string, context string) (string, error) {

	// Truncate diff if too large

	if len(diff) > 15000 {

		diff = diff[:15000] + "\n... [Diff truncated] ..."

	}



	url := "https://api.openai.com/v1/chat/completions"



	systemPrompt := p.SystemPrompt

	if systemPrompt == "" {

		systemPrompt = "You are an expert developer. Generate a raw git commit message. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks."

	}

	

	commitPromptTemplate := p.CommitPrompt

	if commitPromptTemplate == "" {

		commitPromptTemplate = "Generate a git commit message for these changes:\n\n%s\n\n%s"

	}



	userPrompt := fmt.Sprintf(commitPromptTemplate, diff, context)



	reqBody := openAIChatCompletionRequest{

		Model: p.Model,

		Messages: []struct {

			Role    string `json:"role"`

			Content string `json:"content"`

		}{

			{Role: "system", Content: systemPrompt},

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

	req.Header.Set("Authorization", "Bearer "+p.APIKey)



	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {

		return "", err

	}

	defer resp.Body.Close()



	if resp.StatusCode != http.StatusOK {

		var errResp openAIChatCompletionResponse

		body, _ := io.ReadAll(resp.Body)

		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {

			return "", fmt.Errorf("OpenAI API error: %s (Type: %s)", errResp.Error.Message, errResp.Error.Type)

		}

		return "", fmt.Errorf("OpenAI API error: %s", string(body))

	}



	var result openAIChatCompletionResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {

		return "", err

	}



	if len(result.Choices) > 0 {

		return result.Choices[0].Message.Content, nil

	}



	return "", fmt.Errorf("no response from OpenAI")

}


