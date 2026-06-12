package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DB struct {
	client *firestore.Client
}

// NewDB initializes a new Firestore client.
func NewDB(ctx context.Context, projectID, firestoreID string) (*DB, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, firestoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return &DB{client: client}, nil
}

// Close closes the Firestore client connection.
func (db *DB) Close() error {
	return db.client.Close()
}

// --- CHAT METHODS ---

// SaveChat inserts or updates a chat's settings.
func (db *DB) SaveChat(ctx context.Context, chat *Chat) error {
	docID := fmt.Sprintf("%d", chat.ChatID)
	_, err := db.client.Collection("chats").Doc(docID).Set(ctx, chat)
	return err
}

// GetChat fetches a chat by its ID. Returns nil if not found.
func (db *DB) GetChat(ctx context.Context, chatID int64) (*Chat, error) {
	docID := fmt.Sprintf("%d", chatID)
	doc, err := db.client.Collection("chats").Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil // Chat doesn't exist yet
		}
		return nil, err
	}

	var chat Chat
	if err := doc.DataTo(&chat); err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetAllChats fetches all chats that have been registered/configured.
func (db *DB) GetAllChats(ctx context.Context) ([]Chat, error) {
	var chats []Chat
	// We just fetch all chats. For 1-100 chats, this is instantly fast
	// and uses a fraction of a fraction of the daily free tier.
	iter := db.client.Collection("chats").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var c Chat
		if err := doc.DataTo(&c); err != nil {
			continue // Skip malformed documents
		}
		chats = append(chats, c)
	}
	return chats, nil
}

// --- BIRTHDAY METHODS ---

// AddBirthday saves a new birthday. We use an auto-generated ID for the document.
func (db *DB) AddBirthday(ctx context.Context, b *Birthday) error {
	_, _, err := db.client.Collection("birthdays").Add(ctx, b)
	return err
}

// GetBirthdaysForChat returns all birthdays registered in a specific chat.
func (db *DB) GetBirthdaysForChat(ctx context.Context, chatID int64) ([]Birthday, error) {
	var birthdays []Birthday
	iter := db.client.Collection("birthdays").Where("chat_id", "==", chatID).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var b Birthday
		if err := doc.DataTo(&b); err != nil {
			continue
		}
		birthdays = append(birthdays, b)
	}
	return birthdays, nil
}

// GetTodaysBirthdaysForChat returns birthdays matching today's day and month for a specific chat.
func (db *DB) GetTodaysBirthdaysForChat(ctx context.Context, chatID int64, day, month int) ([]Birthday, error) {
	var birthdays []Birthday
	// Firestore allows chaining multiple Where clauses
	iter := db.client.Collection("birthdays").
		Where("chat_id", "==", chatID).
		Where("month", "==", month).
		Where("day", "==", day).
		Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var b Birthday
		if err := doc.DataTo(&b); err != nil {
			continue
		}
		birthdays = append(birthdays, b)
	}
	return birthdays, nil
}
