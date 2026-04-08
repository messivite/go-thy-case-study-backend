package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/thy-case-study-backend/internal/repo"
)

type GeminiProvider struct {
	apiKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{apiKey: apiKey}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) Respond(ctx context.Context, session repo.ChatSession, history []repo.ChatMessage, prompt string) (string, error) {
	if p.apiKey == "" {
		return "", errors.New("gemini api key is not configured")
	}

	// Replace this placeholder with a real Gemini client call.
	return fmt.Sprintf("Gemini response to: %s", prompt), nil
}
