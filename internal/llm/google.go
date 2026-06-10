package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GoogleClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewGoogleClient(apiKey, model string) *GoogleClient {
	if model == "" {
		model = "gemini-pro"
	}
	return &GoogleClient{apiKey: apiKey, model: model, baseURL: "https://generativelanguage.googleapis.com", client: http.DefaultClient}
}

func (c *GoogleClient) Generate(prompt string) (string, error) {
	payload := map[string]any{
		"contents": []map[string]any{{"parts": []map[string]string{{"text": prompt}}}},
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("google status: %s", resp.Status)
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct{ Text string `json:"text"` } `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("google response missing content")
	}
	return strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text), nil
}