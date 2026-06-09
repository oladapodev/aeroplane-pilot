package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		baseURL = "https://api.openai.com"
	}
	return &OpenAIClient{apiKey: apiKey, model: model, baseURL: baseURL, client: http.DefaultClient}
}

func (c *OpenAIClient) Generate(prompt string) (string, error) {
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
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
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai status: %s", resp.Status)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai response missing choices")
	}
	return parsed.Choices[0].Message.Content, nil
}
