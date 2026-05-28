package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (p *GeminiProvider) AskChat(prompt string, contextStr string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.Model, p.APIKey)

	fullPrompt := prompt
	if contextStr != "" {
		fullPrompt = fmt.Sprintf("Context:\n%s\n\nQuestion:\n%s", contextStr, prompt)
	}
	if p.SystemPrompt != "" {
		fullPrompt = p.SystemPrompt + "\n\n" + fullPrompt
	}

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
			return "", fmt.Errorf("Gemini API error: %s", errResp.Error.Message)
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
