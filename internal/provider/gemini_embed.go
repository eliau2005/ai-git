package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type geminiEmbedRequest struct {
	Model   string `json:"model"`
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) GenerateEmbedding(text string) ([]float32, error) {
	// Gemini uses text-embedding-004
	model := "text-embedding-004"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", model, p.APIKey)

	reqBody := geminiEmbedRequest{
		Model: "models/" + model,
	}
	reqBody.Content.Parts = []struct {
		Text string `json:"text"`
	}{
		{Text: text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini Embed API error: %s", string(body))
	}

	var result geminiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s", result.Error.Message)
	}

	return result.Embedding.Values, nil
}
