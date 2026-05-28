package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (p *GeminiProvider) AskChatStream(prompt string, contextStr string, onChunk func(string)) error {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.Model, p.APIKey)

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
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// We use a longer timeout for streaming
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp geminiGenerateContentResponse
		body, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return fmt.Errorf("Gemini API error: %s", errResp.Error.Message)
		}
		return fmt.Errorf("Gemini API error: %s", string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var chunkResp geminiGenerateContentResponse
			if err := json.Unmarshal([]byte(data), &chunkResp); err == nil {
				if len(chunkResp.Candidates) > 0 && len(chunkResp.Candidates[0].Content.Parts) > 0 {
					text := chunkResp.Candidates[0].Content.Parts[0].Text
					onChunk(text)
				}
			}
		}
	}

	return scanner.Err()
}
