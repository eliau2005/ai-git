package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *GeminiProvider) RefactorCode(prompt string, fileContent string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.Model, p.APIKey)

	systemPrompt := "You are an expert autonomous developer. Your goal is to refactor or modify the provided code according to the user's instructions. " +
		"Return ONLY the raw new code for the file. Do not include markdown code blocks (like ```go). " +
		"Do not explain your changes. Output exactly what should be written to the file so it can be saved directly."

	fullPrompt := fmt.Sprintf("%s\n\nUser Prompt: %s\n\nFile Content:\n%s", systemPrompt, prompt, fileContent)

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
					{Text: fullPrompt},
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
		text := result.Candidates[0].Content.Parts[0].Text
		// Cleanup Markdown formatting if AI ignored instructions
		text = strings.TrimPrefix(text, "```go\n")
		text = strings.TrimPrefix(text, "```javascript\n")
		text = strings.TrimPrefix(text, "```ts\n")
		text = strings.TrimPrefix(text, "```typescript\n")
		text = strings.TrimPrefix(text, "```\n")
		text = strings.TrimSuffix(text, "\n```")
		text = strings.TrimSuffix(text, "```")
		return text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}
