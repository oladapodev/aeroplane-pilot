package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AnthropicClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicClient{apiKey: apiKey, model: model, baseURL: "https://api.anthropic.com", client: http.DefaultClient}
}

func (c *AnthropicClient) Generate(prompt string) (string, error) {
	payload := map[string]any{
		"model":      c.model,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens": 512,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.baseURL+"/v1/messages", bytes.NewReader(body))
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic status: %s", resp.Status)
	}
	var parsed struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("anthropic response missing content")
	}
	return parsed.Content[0].Text, nil
}