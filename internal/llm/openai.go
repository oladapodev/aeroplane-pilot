package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIClient struct {
	apiKey string
	model  string
	baseURL string
	client *http.Client
}

func NewOpenAIClient(apiKey, model, baseURL string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenAIClient{apiKey: apiKey, model: model, baseURL: baseURL, client: http.DefaultClient}
}

func (c *OpenAIClient) Generate(prompt string) (string, error) {
	payload := map[string]any{
		"model":      c.model,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 256,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", fmt.Errorf("json decode error: %v, body: %s", err, string(bodyBytes))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai response missing choices, body: %s", string(bodyBytes))
	}
	return parsed.Choices[0].Message.Content, nil
}
