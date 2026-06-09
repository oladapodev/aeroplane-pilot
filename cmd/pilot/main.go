package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	discordgo "github.com/bwmarrin/discordgo"
	telegram "github.com/go-telegram/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	"github.com/oladapodev/aeroplane-pilot/internal/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/config"
	"github.com/oladapodev/aeroplane-pilot/internal/core"
	"github.com/oladapodev/aeroplane-pilot/internal/llm"
)

type telegramStarter func(context.Context, string, bot.Handler) (*telegram.Bot, error)

type discordStarter func(context.Context, string, bot.Handler) (*discordgo.Session, error)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, cfg, nil, startTelegram, startDiscord, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, client aeroplaneClient, startTelegram telegramStarter, startDiscord discordStarter, out io.Writer) error {
	if cfg.LLMAPIKey == "" {
		return fmt.Errorf("missing LLM_API_KEY or OPENAI_API_KEY")
	}

	llmClient, err := llm.NewFromConfig(llm.Config{Provider: cfg.LLMProvider, APIKey: cfg.LLMAPIKey, Model: cfg.LLMModel})
	if err != nil {
		return err
	}

	if client == nil {
		client, err = aeroplane.NewClient(cfg.DBPath)
		if err != nil {
			return err
		}
		defer client.Close()
	}

	router := core.NewRouter(client, core.WithLLM(llmClient))

	if cfg.TelegramToken != "" {
		if _, err := startTelegram(ctx, cfg.TelegramToken, router); err != nil {
			return err
		}
		<-ctx.Done()
		return nil
	}
	if cfg.DiscordToken != "" {
		if _, err := startDiscord(ctx, cfg.DiscordToken, router); err != nil {
			return err
		}
		<-ctx.Done()
		return nil
	}

	_, err = fmt.Fprintln(out, router.Route("/health"))
	return err
}

type aeroplaneClient interface {
	ListServices() ([]aeroplane.Service, error)
	FailedDeployments(limit int) ([]aeroplane.Deployment, error)
	Close() error
}

func startTelegram(ctx context.Context, token string, handler bot.Handler) (*telegram.Bot, error) {
	return bot.StartTelegram(ctx, token, handler)
}

func startDiscord(ctx context.Context, token string, handler bot.Handler) (*discordgo.Session, error) {
	return bot.StartDiscord(ctx, token, handler)
}
