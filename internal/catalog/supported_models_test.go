package catalog

import (
	"testing"

	"github.com/messivite/go-thy-case-study-backend/internal/provider"
)

func TestSupportedModelsFromRegistry_NilRegistry(t *testing.T) {
	if got := SupportedModelsFromRegistry(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestSupportedModelsFromRegistry_OpenAI(t *testing.T) {
	reg := provider.NewRegistry("openai")
	reg.Register(provider.NewOpenAIProvider("sk-test", "gpt-4o"), provider.ProviderMeta{
		Name: "openai", DefaultModel: "gpt-4o", RequiredEnvKey: "OPENAI_API_KEY", SupportsStream: true,
	})
	models := SupportedModelsFromRegistry(reg)
	if len(models) == 0 {
		t.Fatal("expected at least one model from openai template")
	}
}
