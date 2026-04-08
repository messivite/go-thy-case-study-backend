package provider

import (
	"context"
	"testing"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

type stubProvider struct {
	name string
}

func (s stubProvider) Name() string { return s.name }

func (s stubProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	return domain.ProviderResponse{Content: "ok"}, nil
}

func (s stubProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	events := make(chan domain.StreamEvent, 1)
	events <- domain.StreamEvent{Type: domain.EventDone}
	close(events)
	return events, nil
}

func TestRegistryDefaultProvider(t *testing.T) {
	r := NewRegistry("openai")
	r.Register(stubProvider{name: "openai"}, ProviderMeta{Name: "openai"})
	r.Register(stubProvider{name: "gemini"}, ProviderMeta{Name: "gemini"})

	p, err := r.Get("")
	if err != nil {
		t.Fatalf("expected default provider, got error: %v", err)
	}
	if p.Name() != "openai" {
		t.Fatalf("expected openai as default, got %s", p.Name())
	}
}

func TestRegistryUnknownProvider(t *testing.T) {
	r := NewRegistry("openai")
	r.Register(stubProvider{name: "openai"}, ProviderMeta{Name: "openai"})
	if _, err := r.Get("unknown"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegistrySetDefault(t *testing.T) {
	r := NewRegistry("openai")
	r.Register(stubProvider{name: "openai"}, ProviderMeta{Name: "openai"})
	r.Register(stubProvider{name: "gemini"}, ProviderMeta{Name: "gemini"})

	if err := r.SetDefault("gemini"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if r.Default() != "gemini" {
		t.Fatalf("expected gemini as default, got %s", r.Default())
	}
}

func TestRegistryListNames(t *testing.T) {
	r := NewRegistry("openai")
	r.Register(stubProvider{name: "openai"}, ProviderMeta{Name: "openai"})
	r.Register(stubProvider{name: "gemini"}, ProviderMeta{Name: "gemini"})

	names := r.ListNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(names))
	}
}
