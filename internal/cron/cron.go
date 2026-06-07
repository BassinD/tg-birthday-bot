package cron

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/BassinD/tg-birthday-bot/internal/ai"
	"github.com/BassinD/tg-birthday-bot/internal/db"
	"github.com/BassinD/tg-birthday-bot/internal/i18n"
	"github.com/BassinD/tg-birthday-bot/internal/telegram"
)

type Service struct {
	db   *db.DB
	ai   *ai.Client
	bot  *telegram.Bot
	i18n *i18n.Translator
}

// NewService initializes the cron service.
func NewService(database *db.DB, aiClient *ai.Client, tgBot *telegram.Bot, translator *i18n.Translator) *Service {
	return &Service{
		db:   database,
		ai:   aiClient,
		bot:  tgBot,
		i18n: translator,
	}
}

// TriggerHandler is the HTTP endpoint hit by GCP Cloud Scheduler.
func (s *Service) TriggerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// 1. Fetch all registered chats
	chats, err := s.db.GetAllChats(ctx)
	if err != nil {
		log.Printf("Cron DB Error: failed to fetch chats: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, chat := range chats {
		// Skip chats that haven't fully configured their settings
		if chat.CelebrationTime == "" || chat.Timezone == "" {
			log.Printf("Chat %d skipped due to not complited configuration", chat.ChatID)
			continue
		}

		// 2. Load the chat's specific timezone
		loc, err := time.LoadLocation(chat.Timezone)
		if err != nil {
			log.Printf("Cron TZ Error for chat %d: %v", chat.ChatID, err)
			continue
		}

		// 3. What time is it RIGHT NOW in their timezone?
		nowInChatTZ := time.Now().In(loc)

		// Parse their saved time (e.g., "09:00")
		targetTime, err := time.Parse("15:04", chat.CelebrationTime)
		if err != nil {
			continue
		}

		// 4. If the current hour matches their target hour, celebrate!
		// (We match by hour to tolerate slight trigger delays from Cloud Scheduler)
		if nowInChatTZ.Hour() == targetTime.Hour() {
			s.celebrateBirthdays(ctx, &chat, nowInChatTZ)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cron execution complete."))
}

func (s *Service) celebrateBirthdays(ctx context.Context, chat *db.Chat, localTime time.Time) {
	lang := chat.Language
	if lang == "" {
		lang = "ru"
	}

	// 1. Fetch anyone born on this day and month in this specific chat
	birthdays, err := s.db.GetTodaysBirthdaysForChat(ctx, chat.ChatID, localTime.Day(), int(localTime.Month()))
	if err != nil {
		log.Printf("Cron DB Error fetching birthdays for chat %d: %v", chat.ChatID, err)
		return
	}

	// 2. Loop through the birthday boys/girls and send the AI wishes!
	for _, b := range birthdays {
		// Generate the custom wish via Gemini
		wish, err := s.ai.GenerateWish(ctx, b.Username, chat.PromptTemplate)

		if err != nil {
			// Fallback to a hardcoded string if the AI fails or limits are reached
			log.Printf("AI Generation failed for %s: %v", b.Username, err)
			wish = s.i18n.T(lang, "fallback_wish", b.Username)
		}

		// Send it to the chat
		s.bot.SendText(chat.ChatID, wish)
	}
}
