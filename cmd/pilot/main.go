package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	"github.com/oladapodev/aeroplane-pilot/internal/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/config"
	"github.com/oladapodev/aeroplane-pilot/internal/core"
	"github.com/oladapodev/aeroplane-pilot/internal/llm"
)

func main() {
	cfg := config.Load()
	if cfg.LLMAPIKey == "" {
		fmt.Fprintln(os.Stderr, "missing LLM_API_KEY or OPENAI_API_KEY")
		os.Exit(1)
	}

	llmClient, err := llm.NewFromConfig(llm.Config{Provider: cfg.LLMProvider, APIKey: cfg.LLMAPIKey, Model: cfg.LLMModel})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	aClient, err := aeroplane.NewClient(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer aClient.Close()

	router := core.NewRouter(aClient, core.WithLLM(llmClient))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if cfg.TelegramToken != "" {
		if _, err := bot.StartTelegram(ctx, cfg.TelegramToken, router); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		<-ctx.Done()
		return
	}

	fmt.Println(router.Route("/health"))
}
