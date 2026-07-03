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

	log.Printf("📥 Received command '/%s' with args '%s' from chat %d", command, args, chatID)

	// 1. Fetch or initialize chat settings
	chat, err := b.db.GetChat(ctx, chatID)
	if err != nil {
		log.Printf("❌ ERROR fetching chat %d: %v", chatID, err)
		return
	}

	// 2. Set default language to Russian if it's a new chat or not set
	if chat == nil {
		log.Printf("ℹ️ Initializing new chat record for %d", chatID)
		chat = &db.Chat{ChatID: chatID, Language: "ru"}
	} else if chat.Language == "" {
		chat.Language = "ru"
	}

	lang := chat.Language

	var username string
	if msg.From != nil {
		username = msg.From.UserName
	}

	// 3. Route the command
	switch command {
	case "start":
		if err := b.db.SaveChat(ctx, chat); err != nil {
			log.Printf("❌ ERROR saving new chat %d on /start: %v", chatID, err)
		} else {
			log.Printf("✅ Successfully registered chat %d", chatID)
		}
		b.SendText(chatID, b.i18n.T(lang, "start_welcome"))

	case "add_birthday":
		b.handleAddBirthday(ctx, chat, username, args, lang)

	case "list_birthdays":
		b.handleListBirthdays(ctx, chat, lang)

	case "set_time":
		b.handleSetTime(ctx, chat, args, lang)

	case "set_template":
		b.handleSetTemplate(ctx, chat, args, lang)

	case "set_language":
		b.handleSetLanguage(ctx, chat, args)

	case "delete_birthday":
		b.handleDeleteBirthday(ctx, chat, username, lang)

	default:
		log.Printf("⚠️ Unknown command '/%s' from chat %d", command, chatID)
		b.SendText(chatID, b.i18n.T(lang, "err_unknown_command"))
	}
}

func (b *Bot) handleAddBirthday(ctx context.Context, chat *db.Chat, username, args, lang string) {
	if username == "" {
		log.Printf("⚠️ /add_birthday from user without username in chat %d", chat.ChatID)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_no_username"))
		return
	}

	dateStr := strings.TrimSpace(args)
	if dateStr == "" {
		log.Printf("⚠️ Invalid /add_birthday format from chat %d: %s", chat.ChatID, args)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_add_format"))
		return
	}

	dateParts := strings.Split(dateStr, "-")
	if len(dateParts) != 2 {
		log.Printf("⚠️ Invalid date format from chat %d: %s", chat.ChatID, dateStr)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_date_format"))
		return
	}

	day, err1 := strconv.Atoi(dateParts[0])
	month, err2 := strconv.Atoi(dateParts[1])
	if err1 != nil || err2 != nil || day < 1 || day > 31 || month < 1 || month > 12 {
		log.Printf("⚠️ Invalid date logic from chat %d: %s", chat.ChatID, dateStr)
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
		log.Printf("❌ ERROR saving birthday for @%s in chat %d: %v", username, chat.ChatID, err)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	log.Printf("✅ Successfully saved birthday for @%s (%02d-%02d) in chat %d", username, day, month, chat.ChatID)
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_birthday_saved", username, day, month))
}

func (b *Bot) handleListBirthdays(ctx context.Context, chat *db.Chat, lang string) {
	birthdays, err := b.db.GetBirthdaysForChat(ctx, chat.ChatID)
	if err != nil {
		log.Printf("❌ ERROR fetching birthdays for chat %d: %v", chat.ChatID, err)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	if len(birthdays) == 0 {
		log.Printf("ℹ️ No birthdays found for chat %d", chat.ChatID)
		b.SendText(chat.ChatID, b.i18n.T(lang, "empty_birthdays"))
		return
	}

	log.Printf("✅ Listed %d birthdays for chat %d", len(birthdays), chat.ChatID)
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
		log.Printf("⚠️ Invalid /set_time format from chat %d: %s", chat.ChatID, args)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_set_time_format"))
		return
	}

	timeStr, tzStr := parts[0], parts[1]

	if _, err := time.Parse("15:04", timeStr); err != nil {
		log.Printf("⚠️ Invalid time string from chat %d: %s", chat.ChatID, timeStr)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_invalid_time"))
		return
	}

	if _, err := time.LoadLocation(tzStr); err != nil {
		log.Printf("⚠️ Invalid timezone from chat %d: %s", chat.ChatID, tzStr)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_invalid_tz"))
		return
	}

	chat.CelebrationTime = timeStr
	chat.Timezone = tzStr
	if err := b.db.SaveChat(ctx, chat); err != nil {
		log.Printf("❌ ERROR saving time settings for chat %d: %v", chat.ChatID, err)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	log.Printf("✅ Successfully updated time settings for chat %d: %s %s", chat.ChatID, timeStr, tzStr)
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_time_set", timeStr, tzStr))
}

func (b *Bot) handleSetTemplate(ctx context.Context, chat *db.Chat, args, lang string) {
	if strings.TrimSpace(args) == "" {
		log.Printf("⚠️ Empty /set_template from chat %d", chat.ChatID)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_template_empty"))
		return
	}

	chat.PromptTemplate = args
	if err := b.db.SaveChat(ctx, chat); err != nil {
		log.Printf("❌ ERROR saving template for chat %d: %v", chat.ChatID, err)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	log.Printf("✅ Successfully updated AI template for chat %d", chat.ChatID)
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_template_set"))
}

func (b *Bot) handleSetLanguage(ctx context.Context, chat *db.Chat, args string) {
	newLang := strings.ToLower(strings.TrimSpace(args))

	if newLang != "en" && newLang != "ru" {
		log.Printf("⚠️ Invalid language selection from chat %d: %s", chat.ChatID, newLang)
		currentLang := chat.Language
		if currentLang == "" {
			currentLang = "ru"
		}
		b.SendText(chat.ChatID, b.i18n.T(currentLang, "err_lang_format"))
		return
	}

	chat.Language = newLang
	if err := b.db.SaveChat(ctx, chat); err != nil {
		log.Printf("❌ ERROR saving language for chat %d: %v", chat.ChatID, err)
		b.SendText(chat.ChatID, b.i18n.T(chat.Language, "err_save_failed"))
		return
	}

	log.Printf("✅ Successfully updated language to '%s' for chat %d", newLang, chat.ChatID)
	b.SendText(chat.ChatID, b.i18n.T(newLang, "success_lang_set"))
}

func (b *Bot) handleDeleteBirthday(ctx context.Context, chat *db.Chat, username, lang string) {
	if username == "" {
		log.Printf("⚠️ /delete_birthday from user without username in chat %d", chat.ChatID)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_no_username"))
		return
	}

	if err := b.db.DeleteBirthday(ctx, chat.ChatID, username); err != nil {
		log.Printf("❌ ERROR deleting birthday for @%s: %v", username, err)
		b.SendText(chat.ChatID, b.i18n.T(lang, "err_save_failed"))
		return
	}

	log.Printf("✅ Deleted birthday for @%s in chat %d", username, chat.ChatID)
	b.SendText(chat.ChatID, b.i18n.T(lang, "success_birthday_deleted", username))
}
