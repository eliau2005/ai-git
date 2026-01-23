package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GeminiProvider struct {
	APIKey       string
	Model        string
	SystemPrompt string
	CommitPrompt string
}

func (p *GeminiProvider) GetName() string {
	return "gemini"
}

type geminiGenerateContentRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) GenerateCommitMessage(diff string, context string) (string, error) {
	// Truncate diff
	if len(diff) > 15000 {
		diff = diff[:15000] + "\n... [Diff truncated] ..."
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.Model, p.APIKey)

	commitPromptTemplate := p.CommitPrompt
	if commitPromptTemplate == "" {
		commitPromptTemplate = "Generate a raw git commit message for the changes below. Output ONLY the message. Structure: a short title, then a blank line, then a description. No conversational filler, no quotes, no backticks.\n\nChanges:\n%s\n\n%s"
	}

	// Incorporate SystemPrompt into the user prompt for Gemini as it doesn't strictly have a separate system role in the same way (or it's complex to structure for v1beta simple calls)
	systemPrompt := p.SystemPrompt
	if systemPrompt != "" {
		commitPromptTemplate = systemPrompt + "\n\n" + commitPromptTemplate
	}

	prompt := fmt.Sprintf(commitPromptTemplate, diff, context)

	reqBody := geminiGenerateContentRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp geminiGenerateContentResponse
		body, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("Gemini API error: %s (Status: %s)", errResp.Error.Message, errResp.Error.Status)
		}
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var result geminiGenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}
