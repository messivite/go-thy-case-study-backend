package provider

import (
	"context"
	"fmt"
	"time"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{apiKey: apiKey, model: model}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	if p.apiKey == "" {
		return domain.ProviderResponse{}, fmt.Errorf("openai api key is not configured")
	}

	prompt := lastUserContent(req.Messages)
	return domain.ProviderResponse{
		Content: fmt.Sprintf("OpenAI response to: %s", prompt),
		Usage: map[string]any{
			"provider": "openai",
			"model":    p.model,
		},
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai api key is not configured")
	}

	events := make(chan domain.StreamEvent, 8)
	prompt := lastUserContent(req.Messages)
	content := fmt.Sprintf("OpenAI response to: %s", prompt)

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
