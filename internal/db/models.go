package db

// Chat represents the settings for a specific Telegram group or private chat.
type Chat struct {
	ChatID int64 `firestore:"chat_id"`

	// The local time the chat wants to celebrate (e.g., "09:00")
	CelebrationTime string `firestore:"celebration_time"`

	// The IANA timezone identifier (e.g., "Europe/Moscow", "America/New_York")
	Timezone string `firestore:"timezone"`

	// The AI instruction for Gemini.
	PromptTemplate string `firestore:"prompt_template"`

	// Language preference (e.g., "en" or "ru")
	Language string `firestore:"language"`
}

// Birthday represents a single user's birthday linked to a specific chat.
type Birthday struct {
	ChatID    int64  `firestore:"chat_id"`
	Username  string `firestore:"username"` // Without the '@' symbol
	FirstName string `firestore:"first_name"`
	LastName  string `firestore:"last_name"`
	Day       int    `firestore:"day"`
	Month     int    `firestore:"month"`
}

// GetDisplayName returns the user's first and last name if available, otherwise it returns the username prefixed with '@'.
func (b Birthday) GetDisplayName() string {
	name := b.FirstName
	if b.LastName != "" {
		if name != "" {
			name += " "
		}
		name += b.LastName
	}

	if name == "" {
		return "@" + b.Username
	}
	return name
}
