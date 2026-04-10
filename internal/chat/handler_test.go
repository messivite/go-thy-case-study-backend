package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func TestWriteAppError_ModelDiscontinued(t *testing.T) {
	w := httptest.NewRecorder()
	writeAppError(w, domain.ErrModelDiscontinued)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWriteAppError_UserCancelled(t *testing.T) {
	w := httptest.NewRecorder()
	writeAppError(w, domain.ErrUserCancelled)

	if w.Code != 499 {
		t.Fatalf("expected status 499, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", body["error"])
	}
	if errObj["code"] != "generation_cancelled" {
		t.Fatalf("expected generation_cancelled, got %v", errObj["code"])
	}
}

func TestHighlightBounds(t *testing.T) {
	start, end, ok := highlightBounds("Merhaba Docker World", "docker")
	if !ok {
		t.Fatal("expected match")
	}
	if start != 8 || end != 14 {
		t.Fatalf("unexpected bounds: %d-%d", start, end)
	}
}

func TestCollectHighlights(t *testing.T) {
	hs := collectHighlights("go", "Go backend", "Use Go for speed")
	if len(hs) != 2 {
		t.Fatalf("expected 2 highlights, got %d", len(hs))
	}
}

type handlerTestQuotaStub struct{}

func (handlerTestQuotaStub) GetUserQuota(context.Context, string) (domain.UserQuota, error) {
	return domain.UserQuota{QuotaBypass: true}, nil
}
func (handlerTestQuotaStub) GetUserTokenUsage(context.Context, string) (domain.UserTokenUsage, error) {
	return domain.UserTokenUsage{}, nil
}
func (handlerTestQuotaStub) FailPendingLog(context.Context, string, string, string, int) error {
	return nil
}
func (handlerTestQuotaStub) CancelPendingLog(context.Context, string) error           { return nil }
func (handlerTestQuotaStub) SetUsageLog(context.Context, string, int, int, int) error { return nil }

func TestHandler_ListModels_unauthorized(t *testing.T) {
	reg := provider.NewRegistry("openai")
	reg.Register(provider.NewOpenAIProvider("k", "gpt-4o"), provider.ProviderMeta{
		Name: "openai", DefaultModel: "gpt-4o", RequiredEnvKey: "OPENAI_API_KEY", SupportsStream: true,
	})
	mem := repo.NewMemoryRepository()
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	h.ListModels(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_ListModels_ok(t *testing.T) {
	ctx := context.Background()
	reg := provider.NewRegistry("openai")
	reg.Register(provider.NewOpenAIProvider("k", "gpt-4o"), provider.ProviderMeta{
		Name: "openai", DefaultModel: "gpt-4o", RequiredEnvKey: "OPENAI_API_KEY", SupportsStream: true,
	})
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "openai", ModelID: "gpt-4o", DisplayName: "GPT-4o", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	req = req.WithContext(auth.ContextWithAuthenticatedUser(ctx, &auth.AuthenticatedUser{UserID: "u1"}))
	h.ListModels(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var body struct {
		Models []struct {
			Provider       string `json:"provider"`
			Model          string `json:"model"`
			DisplayName    string `json:"displayName"`
			SupportsStream bool   `json:"supportsStream"`
		} `json:"models"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Models) != 1 || body.Models[0].Model != "gpt-4o" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
