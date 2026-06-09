package config

import "os"

type Config struct {
	AeroplaneHome   string
	DBPath          string
	LLMProvider     string
	LLMAPIKey       string
	LLMModel        string
	TelegramToken   string
	DiscordToken    string
	WhatsAppToken   string
	DailyHealthTime string
}

func Load() Config {
	aeroplaneHome := getenv("AEROPLANE_HOME", "/opt/aeroplane")
	return Config{
		AeroplaneHome:   aeroplaneHome,
		DBPath:          aeroplaneHome + "/source/data/aeroplane.db",
		LLMProvider:     getenv("LLM_PROVIDER", "openai"),
		LLMAPIKey:       firstNonEmpty(os.Getenv("LLM_API_KEY"), os.Getenv("OPENAI_API_KEY")),
		LLMModel:        getenv("LLM_MODEL", "gpt-4o"),
		TelegramToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		DiscordToken:    os.Getenv("DISCORD_BOT_TOKEN"),
		WhatsAppToken:   os.Getenv("WHATSAPP_BOT_TOKEN"),
		DailyHealthTime: getenv("DAILY_HEALTH_TIME", "08:00"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
