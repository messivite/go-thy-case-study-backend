package provider

import (
	"context"
	"fmt"

	"github.com/example/thy-case-study-backend/internal/repo"
)

type LLMProvider interface {
	Name() string
	Respond(ctx context.Context, session repo.ChatSession, history []repo.ChatMessage, prompt string) (string, error)
}

type ProviderFactory struct {
	providers map[string]LLMProvider
}

func NewProviderFactory(providers []LLMProvider) *ProviderFactory {
	factory := &ProviderFactory{providers: make(map[string]LLMProvider)}
	for _, p := range providers {
		factory.providers[p.Name()] = p
	}
	return factory
}

func (f *ProviderFactory) GetProvider(name string) (LLMProvider, error) {
	if name == "" {
		name = "openai"
	}
	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return provider, nil
}

func (f *ProviderFactory) ListProviders() []string {
	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}
