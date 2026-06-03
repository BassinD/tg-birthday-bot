package ai

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

//go:embed prompt.md
var systemPromptTemplate string

const defaultModel = "gemini-1.5-flash"

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

// GenerateWish calls Gemini to generate a customized birthday message.
func (c *Client) GenerateWish(ctx context.Context, username, stylePrompt string) (string, error) {
	if strings.TrimSpace(stylePrompt) == "" {
		stylePrompt = "Make it a warm, friendly, and cheerful birthday wish."
	}

	// 1. Validate the model resource exists using the corrected signature
	_, err := c.genaiClient.Models.Get(ctx, defaultModel, nil)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve model %s: %w", defaultModel, err)
	}

	// 2. Inject parameters into our embedded markdown prompt
	systemInstruction := fmt.Sprintf(systemPromptTemplate, username, stylePrompt)

	// 3. Generate the actual content
	resp, err := c.genaiClient.Models.GenerateContent(ctx, defaultModel, genai.Text(systemInstruction), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract the text response safely
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return resp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("unexpected response format from gemini")
}
