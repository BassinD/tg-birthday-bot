package ai

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"
)

//go:embed prompt.md
var systemPromptTemplate string

const defaultModel = "gemini-3.5-flash"

type Client struct {
	genaiClient *genai.Client
}

// NewClient initializes the Gemini API client.
func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &Client{genaiClient: client}, nil
}

// GenerateWish calls Gemini to generate a customized birthday message with exponential backoff for resilience.
func (c *Client) GenerateWish(ctx context.Context, name, stylePrompt string) (string, error) {
	if strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Make it a warm, friendly, and cheerful birthday wish."
	}

	// 1. Validate the model resource exists using the corrected signature
	_, err := c.genaiClient.Models.Get(ctx, defaultModel, nil)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve model %s: %w", defaultModel, err)
	}

	// 2. Inject parameters into our embedded markdown prompt
	systemInstruction := fmt.Sprintf(systemPromptTemplate, name, stylePrompt)

	var resp *genai.GenerateContentResponse

	// 3. The Exponential Backoff Loop
	maxRetries := 4
	delay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		// Generate the actual content
		resp, err = c.genaiClient.Models.GenerateContent(ctx, defaultModel, genai.Text(systemInstruction), nil)

		// If successful, break out of the loop immediately
		if err == nil {
			break
		}

		// Inspect the error message to see if it is a transient 503 (Overloaded) or 429 (Rate Limit) error
		errStr := err.Error()
		if strings.Contains(errStr, "503") || strings.Contains(strings.ToLower(errStr), "unavailable") {
			log.Printf("⚠️ Gemini API busy or overloaded. Retrying in %v... (Attempt %d/%d)", delay, i+1, maxRetries)

			// Sleep cleanly, respecting context cancellation if the request is terminated early
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}

			delay *= 2 // Progressively double the wait time: 2s, 4s, 8s...
			continue
		}

		// If it's a completely different error (e.g., Auth, Invalid Argument), do not retry
		break
	}

	// If it still fails after all retries, return the final error wrapped cleanly
	if err != nil {
		return "", fmt.Errorf("failed to generate content after retries: %w", err)
	}

	// 4. Extract the text response safely
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return resp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("unexpected response format from gemini")
}
