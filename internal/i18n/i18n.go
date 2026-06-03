package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

//go:embed locales/*.json
var localeFS embed.FS

// Translator holds all our localized strings in memory.
// Structure: map[languageCode]map[translationKey]translatedString
type Translator struct {
	messages map[string]map[string]string
	fallback string
}

// NewTranslator initializes the localization engine.
func NewTranslator(fallbackLang string) *Translator {
	t := &Translator{
		messages: make(map[string]map[string]string),
		fallback: fallbackLang,
	}

	// Read files from the embedded filesystem
	files, err := localeFS.ReadDir("locales")
	if err != nil {
		log.Fatalf("Failed to read locales directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(file.Name(), ".json")
		data, err := localeFS.ReadFile("locales/" + file.Name())
		if err != nil {
			log.Fatalf("Failed to read locale file %s: %v", file.Name(), err)
		}

		var msgs map[string]string
		if err := json.Unmarshal(data, &msgs); err != nil {
			log.Fatalf("Failed to parse locale %s: %v", file.Name(), err)
		}

		t.messages[langCode] = msgs
	}

	return t
}

// T returns the translated string. It supports formatting via args (like fmt.Sprintf).
func (t *Translator) T(lang, key string, args ...any) string {
	// Fallback to english if the requested language doesn't exist
	langDict, ok := t.messages[lang]
	if !ok {
		langDict = t.messages[t.fallback]
	}

	// Fallback to english if the specific key is missing in the requested language
	msg, ok := langDict[key]
	if !ok {
		msg = t.messages[t.fallback][key]
	}

	// If it's STILL missing (e.g. developer typo), return the key itself so it's obvious in the app
	if msg == "" {
		return "[" + key + "]"
	}

	// Apply formatting if arguments are provided
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
