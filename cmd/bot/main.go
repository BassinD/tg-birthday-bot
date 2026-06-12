package main

import (
	"context"
	"log"
	"net/http"

	"github.com/BassinD/tg-birthday-bot/internal/ai"
	"github.com/BassinD/tg-birthday-bot/internal/config"
	"github.com/BassinD/tg-birthday-bot/internal/cron"
	"github.com/BassinD/tg-birthday-bot/internal/db"
	"github.com/BassinD/tg-birthday-bot/internal/i18n"
	"github.com/BassinD/tg-birthday-bot/internal/telegram"
)

func main() {
	ctx := context.Background()

	// 1. Load Configuration (from .env or System Env)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// 2. Initialize Database (Firestore)
	database, err := db.NewDB(ctx, cfg.GCPProjectID, cfg.FirestoreDBID)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close()

	// 3. Initialize Translations (Fallback to Russian)
	translator := i18n.NewTranslator("ru")
	ix
	// 4. Initialize AI Client (Gemini)
	aiClient, err := ai.NewClient(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("❌ Failed to initialize AI client: %v", err)
	}

	// 5. Initialize Telegram Bot
	bot, err := telegram.NewBot(cfg.TelegramToken, database, translator)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Telegram bot: %v", err)
	}

	// 6. Initialize Cron Service
	cronService := cron.NewService(database, aiClient, bot, translator)

	// 7. Map HTTP Routes
	http.HandleFunc("/webhook", bot.WebhookHandler)
	http.HandleFunc("/cron", cronService.TriggerHandler)

	// A basic health check endpoint (Cloud Run loves this)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Birthday Bot is running! 🥳"))
	})

	// 8. Start the Server
	log.Printf("🚀 Starting server on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
