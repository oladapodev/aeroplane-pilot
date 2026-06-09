package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	discordgo "github.com/bwmarrin/discordgo"
	telegram "github.com/go-telegram/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	"github.com/oladapodev/aeroplane-pilot/internal/bot"
	"github.com/oladapodev/aeroplane-pilot/internal/config"
)

type stubTelegramStarter struct{ called bool }

func (s *stubTelegramStarter) start(_ context.Context, _ string, _ bot.Handler) (*telegram.Bot, error) {
	s.called = true
	return nil, nil
}

type stubDiscordStarter struct{ called bool }

func (s *stubDiscordStarter) start(_ context.Context, _ string, _ bot.Handler) (*discordgo.Session, error) {
	s.called = true
	return nil, nil
}

type fakeClient struct{}

func (fakeClient) ListServices() ([]aeroplane.Service, error) { return []aeroplane.Service{{Status: "running"}}, nil }
func (fakeClient) FailedDeployments(int) ([]aeroplane.Deployment, error) { return nil, nil }
func (fakeClient) Close() error { return nil }

func TestRunUsesTelegramWhenTokenPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	telegramStarter := &stubTelegramStarter{}
	discordStarter := &stubDiscordStarter{}
	cfg := config.Config{LLMAPIKey: "key", TelegramToken: "tg", DiscordToken: "dc"}

	go func() {
		_ = run(ctx, cfg, fakeClient{}, func(ctx context.Context, token string, handler bot.Handler) (*telegram.Bot, error) {
			telegramStarter.called = true
			cancel()
			return nil, nil
		}, func(context.Context, string, bot.Handler) (*discordgo.Session, error) {
			discordStarter.called = true
			return nil, nil
		}, &bytes.Buffer{})
	}()
	<-ctx.Done()
	if !telegramStarter.called {
		t.Fatal("telegram starter was not called")
	}
	if discordStarter.called {
		t.Fatal("discord starter should not be called when telegram token is present")
	}
}

func TestRunUsesDiscordWhenTelegramMissing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	telegramStarter := &stubTelegramStarter{}
	discordStarter := &stubDiscordStarter{}
	cfg := config.Config{LLMAPIKey: "key", DiscordToken: "dc"}

	go func() {
		_ = run(ctx, cfg, fakeClient{}, func(context.Context, string, bot.Handler) (*telegram.Bot, error) {
			telegramStarter.called = true
			return nil, nil
		}, func(ctx context.Context, token string, handler bot.Handler) (*discordgo.Session, error) {
			discordStarter.called = true
			cancel()
			return nil, nil
		}, &bytes.Buffer{})
	}()
	<-ctx.Done()
	if telegramStarter.called {
		t.Fatal("telegram starter should not be called when telegram token is missing")
	}
	if !discordStarter.called {
		t.Fatal("discord starter was not called")
	}
}

func TestRunPrintsHealthWhenNoRealtimeTokens(t *testing.T) {
	cfg := config.Config{LLMAPIKey: "key"}
	var out bytes.Buffer

	if err := run(context.Background(), cfg, fakeClient{}, func(context.Context, string, bot.Handler) (*telegram.Bot, error) { return nil, errors.New("unexpected") }, func(context.Context, string, bot.Handler) (*discordgo.Session, error) { return nil, errors.New("unexpected") }, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if out.String() != "health: 1 services, 1 healthy, 0 unhealthy, 0 failed deployments\n" {
		t.Fatalf("output = %q", out.String())
	}
}
