package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/thy-case-study-backend/internal/repo"
)

type OpenAIProvider struct {
	apiKey string
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{apiKey: apiKey}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Respond(ctx context.Context, session repo.ChatSession, history []repo.ChatMessage, prompt string) (string, error) {
	if p.apiKey == "" {
		return "", errors.New("openai api key is not configured")
	}

	// Replace this placeholder with a real OpenAI client call.
	return fmt.Sprintf("OpenAI response to: %s", prompt), nil
}
