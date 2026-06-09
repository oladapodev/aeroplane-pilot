package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIGenerator(t *testing.T) {
	var receivedAuth string
	var receivedModel string
	var receivedPrompt string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		var payload struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		receivedModel = payload.Model
		if len(payload.Messages) == 0 {
			t.Fatal("expected at least one message")
		}
		receivedPrompt = payload.Messages[len(payload.Messages)-1].Content
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"reply text"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient("secret-key", "gpt-4o", server.URL)
	got, err := client.Generate("hello pilot")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if got != "reply text" {
		t.Fatalf("Generate() = %q, want %q", got, "reply text")
	}
	if receivedAuth != "Bearer secret-key" {
		t.Fatalf("Authorization = %q, want Bearer secret-key", receivedAuth)
	}
	if receivedModel != "gpt-4o" {
		t.Fatalf("model = %q, want gpt-4o", receivedModel)
	}
	if !strings.Contains(receivedPrompt, "hello pilot") {
		t.Fatalf("prompt = %q, want to contain hello pilot", receivedPrompt)
	}
}
