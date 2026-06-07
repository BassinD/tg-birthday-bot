package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/BassinD/tg-birthday-bot/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	command := msg.Command()
	args := msg.CommandArguments()

	// 1. Fetch or initialize chat settings
	chat, err := b.db.GetChat(ctx, chatID)
	if err != nil {
		log.Printf("Error fetching chat: %v", err)
		return
	}

	// 2. Set default language to Russian if it's a new chat or not set
	if chat == nil {
		chat = &db.Chat{ChatID: chatID, Language: "ru"}
	} else if chat.Language == "" {
		chat.Language = "ru"
	}

	lang := chat.Language

	// 3. Route the command
	switch command {
	case "start":
		_ = b.db.SaveChat(ctx, chat)
		b.SendText(chatID, b.i18n.T(lang, "start_welcome"))

	case "add_birthday":
		b.handleAddBirthday(ctx, chat, args, lang)

	case "list_birthdays":
		b.handleListBirthdays(ctx, chat, lang)

	case "set_time":
		b.handleSetTime(ctx, chat, args, lang)

	case "set_template":
		b.handleSetTemplate(ctx, chat, args, lang)

	case "set_language":
		b.handleSetLanguage(ctx, chat, args) // We handle lang inside, as it might change!

	default:
		b.SendText(chatID, b.i18n.T(lang, "err_unknown_command"))
	}
}

func (b *Bot) handleAddBirthday(ctx context.Context, chat *db.Chat, args, lang string) {
	parts := strings.Fields(args)
	if len(parts) != 2 {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_add_format"))
		return
	}

	username := strings.TrimPrefix(parts[0], "@")
	dateParts := strings.Split(parts[1], "-")
	if len(dateParts) != 2 {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_date_format"))
		return
	}

	day, err1 := strconv.Atoi(dateParts[0])
	month, err2 := strconv.Atoi(dateParts[1])
	if err1 != nil || err2 != nil || day < 1 || day > 31 || month < 1 || month > 12 {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_invalid_date"))
		return
	}

	birthday := &db.Birthday{
		ChatID:   chat.ChatID,
		Username: username,
		Day:      day,
		Month:    month,
	}

	if err := b.db.AddBirthday(ctx, birthday); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	// Using T() with formatting arguments
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_birthday_saved", username, day, month))
}

func (b *Bot) handleListBirthdays(ctx context.Context, chat *db.Chat, lang string) {
	birthdays, err := b.db.GetBirthdaysForChat(ctx, chat.ChatID)
	if err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	if len(birthdays) == 0 {
		b.SendText(chat.ChatID, b.i18n.T(lang, "empty_birthdays"))
		return
	}

	var msg strings.Builder
	msg.WriteString(b.i18n.T(lang, "list_title") + "\n")
	for _, b := range birthdays {
		msg.WriteString(fmt.Sprintf("• @%s: %02d-%02d\n", b.Username, b.Day, b.Month))
	}
	b.SendText(chat.ChatID, msg.String())
}

func (b *Bot) handleSetTime(ctx context.Context, chat *db.Chat, args, lang string) {
	parts := strings.Fields(args)
	if len(parts) != 2 {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_set_time_format"))
		return
	}

	timeStr, tzStr := parts[0], parts[1]

	if _, err := time.Parse("15:04", timeStr); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_invalid_time"))
		return
	}

	if _, err := time.LoadLocation(tzStr); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_invalid_tz"))
		return
	}

	chat.CelebrationTime = timeStr
	chat.Timezone = tzStr
	if err := b.db.SaveChat(ctx, chat); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_time_set", timeStr, tzStr))
}

func (b *Bot) handleSetTemplate(ctx context.Context, chat *db.Chat, args, lang string) {
	if strings.TrimSpace(args) == "" {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_template_empty"))
		return
	}

	chat.PromptTemplate = args
	if err := b.db.SaveChat(ctx, chat); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_template_set"))
}

func (b *Bot) handleSetLanguage(ctx context.Context, chat *db.Chat, args string) {
	newLang := strings.ToLower(strings.TrimSpace(args))

	if newLang != "en" && newLang != "ru" {
		// Use the current language to tell them they messed up the command
		currentLang := chat.Language
		if currentLang == "" {
			currentLang = "ru"
		}
		b.SendText(chat.ChatID, b.i18n.T(currentLang, "err_lang_format"))
		return
	}

	chat.Language = newLang
	if err := b.db.SaveChat(ctx, chat); err != nil {
		b.SendText(chat.ChatID, b.i18n.T(chat.Language, "err_save_failed"))
		return
	}

	// Reply in the newly selected language!
	b.SendText(chat.ChatID, b.i18n.T(newLang, "success_lang_set"))
}
