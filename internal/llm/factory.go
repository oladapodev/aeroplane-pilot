package llm

import "fmt"

type Config struct {
	Provider string
	APIKey   string
	Model    string
}

func NewFromConfig(cfg Config) (Generator, error) {
	switch cfg.Provider {
	case "", "openai":
		return NewOpenAIClient(cfg.APIKey, cfg.Model, ""), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}
}
