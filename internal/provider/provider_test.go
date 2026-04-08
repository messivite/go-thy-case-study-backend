package provider

import (
	"context"
	"testing"

	"github.com/example/thy-case-study-backend/internal/repo"
)

type stubProvider struct {
	name string
}

func (s stubProvider) Name() string { return s.name }

func (s stubProvider) Respond(ctx context.Context, session repo.ChatSession, history []repo.ChatMessage, prompt string) (string, error) {
	return "ok", nil
}

func TestProviderFactoryDefaultProvider(t *testing.T) {
	f := NewProviderFactory([]LLMProvider{
		stubProvider{name: "openai"},
		stubProvider{name: "gemini"},
	})

	p, err := f.GetProvider("")
	if err != nil {
		t.Fatalf("expected default provider, got error: %v", err)
	}
	if p.Name() != "openai" {
		t.Fatalf("expected openai as default, got %s", p.Name())
	}
}

func TestProviderFactoryUnknownProvider(t *testing.T) {
	f := NewProviderFactory([]LLMProvider{stubProvider{name: "openai"}})
	if _, err := f.GetProvider("unknown"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
