package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	envTelegramToken = "TELEGRAM_BOT_TOKEN"
	envGeminiAPIKey  = "GEMINI_API_KEY"
	envGCPProjectID  = "GCP_PROJECT_ID"
	envPort          = "PORT"
	envDailyLimit    = "DAILY_MSG_LIMIT"
)

// Config holds all the environment variables required for the application.
type Config struct {
	TelegramToken string
	GeminiAPIKey  string
	GCPProjectID  string
	Port          string
	DailyMsgLimit int
}

// Load reads environment variables and returns a Config struct.
// In local development, make sure you export these or run the app with a .env loader.
func Load() (*Config, error) {
	// Parse the daily limit, default to 5 if not set or invalid
	limitStr := os.Getenv(envDailyLimit)
	limit := 5
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	cfg := &Config{
		TelegramToken: os.Getenv(envTelegramToken),
		GeminiAPIKey:  os.Getenv(envGeminiAPIKey),
		GCPProjectID:  os.Getenv(envGCPProjectID),
		Port:          os.Getenv(envPort),
		DailyMsgLimit: limit,
	}

	// Validation
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("%s is not set", envTelegramToken)
	}
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("%s is not set", envGeminiAPIKey)
	}
	if cfg.GCPProjectID == "" {
		return nil, fmt.Errorf("%s is not set", envGCPProjectID)
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
