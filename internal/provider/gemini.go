package provider

import (
	"context"
	"fmt"
	"time"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

type GeminiProvider struct {
	apiKey string
	model  string
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &GeminiProvider{apiKey: apiKey, model: model}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	if p.apiKey == "" {
		return domain.ProviderResponse{}, fmt.Errorf("gemini api key is not configured")
	}

	prompt := lastUserContent(req.Messages)
	return domain.ProviderResponse{
		Content: fmt.Sprintf("Gemini response to: %s", prompt),
		Usage: map[string]any{
			"provider": "gemini",
			"model":    p.model,
		},
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini api key is not configured")
	}

	events := make(chan domain.StreamEvent, 8)
	prompt := lastUserContent(req.Messages)
	content := fmt.Sprintf("Gemini response to: %s", prompt)

	go func() {
		defer close(events)
		for _, chunk := range splitChunks(content, 24) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(40 * time.Millisecond):
				events <- domain.StreamEvent{Type: domain.EventDelta, Delta: chunk}
			}
		}
		events <- domain.StreamEvent{Type: domain.EventDone}
	}()

	return events, nil
}
