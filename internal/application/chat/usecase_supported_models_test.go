package chat

import (
	"context"
	"errors"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func registryWithOpenAI(t *testing.T) *provider.Registry {
	t.Helper()
	reg := provider.NewRegistry("openai")
	reg.Register(provider.NewOpenAIProvider("sk-test", "gpt-4o"), provider.ProviderMeta{
		Name: "openai", DefaultModel: "gpt-4o", RequiredEnvKey: "OPENAI_API_KEY", SupportsStream: true,
	})
	return reg
}

func quotaBypass() domain.QuotaRepository {
	return &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}
}

func TestUseCase_ListSupportedModels_fromCatalog(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "openai", ModelID: "custom", DisplayName: "Custom", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	uc := NewUseCase(mem, quotaBypass(), registryWithOpenAI(t), mem)
	got, err := uc.ListSupportedModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ModelID != "custom" {
		t.Fatalf("got %+v", got)
	}
}

func TestUseCase_ListSupportedModels_nilCatalogUsesRegistry(t *testing.T) {
	mem := repo.NewMemoryRepository()
	uc := NewUseCase(mem, quotaBypass(), registryWithOpenAI(t), nil)
	got, err := uc.ListSupportedModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected registry-derived models")
	}
}

func TestUseCase_CreateSession_ErrModelDiscontinued(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "openai", ModelID: "only-this", DisplayName: "X", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	uc := NewUseCase(mem, quotaBypass(), registryWithOpenAI(t), mem)
	_, err := uc.CreateSession(ctx, "u1", "t1", "openai", "gpt-4o")
	if !errors.Is(err, domain.ErrModelDiscontinued) {
		t.Fatalf("want ErrModelDiscontinued, got %v", err)
	}
}

func TestUseCase_CreateSession_modelAllowed(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "openai", ModelID: "gpt-4o", DisplayName: "GPT-4o", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	uc := NewUseCase(mem, quotaBypass(), registryWithOpenAI(t), mem)
	s, err := uc.CreateSession(ctx, "u1", "t1", "openai", "gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if s.DefaultProvider != "openai" || s.DefaultModel != "gpt-4o" {
		t.Fatalf("session defaults: %+v", s)
	}
}
