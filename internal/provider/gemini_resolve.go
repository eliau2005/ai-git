package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *GeminiProvider) ResolveConflict(fileContent string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.Model, p.APIKey)

	systemPrompt := "You are an expert developer resolving git merge conflicts. " +
		"I will provide you with a file containing standard git conflict markers (<<<<<<<, =======, >>>>>>>). " +
		"Your job is to understand the context of the conflicting changes and output the fully merged file without any conflict markers. " +
		"Return ONLY the raw code for the resolved file. Do not include markdown code blocks (like ```go). Output exactly what should be written to the file."

	prompt := fmt.Sprintf("%s\n\nFile Content:\n%s", systemPrompt, fileContent)

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
		text := result.Candidates[0].Content.Parts[0].Text
		text = strings.TrimPrefix(text, "```go\n")
		text = strings.TrimPrefix(text, "```\n")
		text = strings.TrimSuffix(text, "\n```")
		text = strings.TrimSuffix(text, "```")
		return text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}
