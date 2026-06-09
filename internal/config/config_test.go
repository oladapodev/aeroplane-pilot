package config

import "testing"

func TestLoadDefaultConfig(t *testing.T) {
	t.Setenv("AEROPLANE_HOME", "")
	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("LLM_MODEL", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("DISCORD_BOT_TOKEN", "")
	t.Setenv("WHATSAPP_BOT_TOKEN", "")
	t.Setenv("DAILY_HEALTH_TIME", "")

	cfg := Load()

	if cfg.AeroplaneHome != "/opt/aeroplane" {
		t.Fatalf("AeroplaneHome = %q, want /opt/aeroplane", cfg.AeroplaneHome)
	}
	if cfg.LLMProvider != "openai" {
		t.Fatalf("LLMProvider = %q, want openai", cfg.LLMProvider)
	}
	if cfg.LLMModel != "gpt-4o" {
		t.Fatalf("LLMModel = %q, want gpt-4o", cfg.LLMModel)
	}
	if cfg.DailyHealthTime != "08:00" {
		t.Fatalf("DailyHealthTime = %q, want 08:00", cfg.DailyHealthTime)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv("AEROPLANE_HOME", "/srv/aeroplane")
	t.Setenv("LLM_PROVIDER", "anthropic")
	t.Setenv("LLM_API_KEY", "llm-key")
	t.Setenv("LLM_MODEL", "claude-3-5-sonnet")
	t.Setenv("TELEGRAM_BOT_TOKEN", "telegram-token")
	t.Setenv("DISCORD_BOT_TOKEN", "discord-token")
	t.Setenv("WHATSAPP_BOT_TOKEN", "whatsapp-token")
	t.Setenv("DAILY_HEALTH_TIME", "09:30")

	cfg := Load()

	if cfg.AeroplaneHome != "/srv/aeroplane" {
		t.Fatalf("AeroplaneHome = %q, want /srv/aeroplane", cfg.AeroplaneHome)
	}
	if cfg.LLMProvider != "anthropic" {
		t.Fatalf("LLMProvider = %q, want anthropic", cfg.LLMProvider)
	}
	if cfg.LLMAPIKey != "llm-key" {
		t.Fatalf("LLMAPIKey = %q, want llm-key", cfg.LLMAPIKey)
	}
	if cfg.LLMModel != "claude-3-5-sonnet" {
		t.Fatalf("LLMModel = %q, want claude-3-5-sonnet", cfg.LLMModel)
	}
	if cfg.TelegramToken != "telegram-token" || cfg.DiscordToken != "discord-token" || cfg.WhatsAppToken != "whatsapp-token" {
		t.Fatalf("token overrides not loaded: %#v", cfg)
	}
	if cfg.DailyHealthTime != "09:30" {
		t.Fatalf("DailyHealthTime = %q, want 09:30", cfg.DailyHealthTime)
	}
}
