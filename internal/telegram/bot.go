package telegram

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/BassinD/tg-birthday-bot/internal/db"
	"github.com/BassinD/tg-birthday-bot/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api  *tgbotapi.BotAPI
	db   *db.DB
	i18n *i18n.Translator
}

// NewBot initializes the Telegram bot client.
func NewBot(token string, database *db.DB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	// Uncomment the next line if you want verbose logging from the Telegram API during local dev
	// api.Debug = true

	return &Bot{
		api: api,
		db:  database,
	}, nil
}

// WebhookHandler processes incoming POST requests from Telegram.
func (b *Bot) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Failed to decode incoming update: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// We only care about messages with commands
	if update.Message == nil || !update.Message.IsCommand() {
		w.WriteHeader(http.StatusOK) // Always return 200 OK to Telegram, otherwise they retry
		return
	}

	// Process the command
	ctx := context.Background()
	b.handleCommand(ctx, update.Message)

	w.WriteHeader(http.StatusOK)
}

// sendText is a helper to easily reply to a chat.
func (b *Bot) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send message to %d: %v", chatID, err)
	}
}
