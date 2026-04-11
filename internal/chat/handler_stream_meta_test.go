package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

type sseHappyStreamProvider struct{}

func (sseHappyStreamProvider) Name() string { return "ssehappy" }

func (sseHappyStreamProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_, _ = ctx, req
	return domain.ProviderResponse{Content: "x"}, nil
}

func (sseHappyStreamProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_, _ = ctx, req
	ch := make(chan domain.StreamEvent, 1)
	ch <- domain.StreamEvent{Type: domain.EventDelta, Delta: "ok"}
	close(ch)
	return ch, nil
}

// one delta then channel left open — client context cancel triggers partial path
type sseHangStreamProvider struct{}

func (sseHangStreamProvider) Name() string { return "ssehang" }

func (sseHangStreamProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_, _ = ctx, req
	return domain.ProviderResponse{Content: "x"}, nil
}

func (sseHangStreamProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_, _ = ctx, req
	ch := make(chan domain.StreamEvent, 1)
	ch <- domain.StreamEvent{Type: domain.EventDelta, Delta: "partial"}
	return ch, nil
}

func parseSSEEvents(body string) []map[string]any {
	var out []map[string]any
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" || strings.HasPrefix(block, ":") {
			continue
		}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			raw := strings.TrimPrefix(line, "data: ")
			var m map[string]any
			if json.Unmarshal([]byte(raw), &m) == nil {
				out = append(out, m)
			}
		}
	}
	return out
}

func TestHandler_StreamMessage_openingMetaHasBothIDsAndSingleMetaEvent(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "ssehappy", ModelID: "m1", DisplayName: "M1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	s, err := mem.CreateChatSession(ctx, "u1", "t", "ssehappy", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("ssehappy")
	reg.Register(sseHappyStreamProvider{}, provider.ProviderMeta{
		Name: "ssehappy", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := auth.ContextWithAuthenticatedUser(req.Context(), &auth.AuthenticatedUser{UserID: "u1"})
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	r.Post("/api/chats/{chatID}/stream", h.StreamMessage)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/chats/"+s.ID.String()+"/stream", strings.NewReader(`{"content":"hi","provider":"ssehappy","model":"m1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rr.Code, rr.Body.String())
	}
	events := parseSSEEvents(rr.Body.String())
	metaCount := 0
	var firstMeta map[string]any
	for _, ev := range events {
		if ev["type"] != "meta" {
			continue
		}
		metaCount++
		meta, _ := ev["meta"].(map[string]any)
		if meta == nil {
			continue
		}
		if firstMeta == nil {
			firstMeta = meta
		}
	}
	if metaCount != 1 {
		t.Fatalf("expected exactly one meta event, got %d: %v", metaCount, events)
	}
	if firstMeta == nil {
		t.Fatalf("no meta: %v", events)
	}
	if firstMeta["userMessageId"] == nil || firstMeta["userMessageId"] == "" {
		t.Fatalf("first meta: %+v", firstMeta)
	}
	if firstMeta["assistantMessageId"] == nil || firstMeta["assistantMessageId"] == "" {
		t.Fatalf("first meta missing assistantMessageId: %+v", firstMeta)
	}
	var hasDone bool
	for _, ev := range events {
		if ev["type"] == "done" {
			hasDone = true
			break
		}
	}
	if !hasDone {
		t.Fatalf("expected type done: %v", events)
	}
}

func TestHandler_StreamMessage_cancelPartialSendsFlagWithoutSecondMeta(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "ssehang", ModelID: "m1", DisplayName: "M1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	s, err := mem.CreateChatSession(ctx, "u1", "t", "ssehang", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("ssehang")
	reg.Register(sseHangStreamProvider{}, provider.ProviderMeta{
		Name: "ssehang", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := auth.ContextWithAuthenticatedUser(req.Context(), &auth.AuthenticatedUser{UserID: "u1"})
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	r.Post("/api/chats/{chatID}/stream", h.StreamMessage)

	reqCtx, cancel := context.WithCancel(context.Background())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/chats/"+s.ID.String()+"/stream", strings.NewReader(`{"content":"hi","provider":"ssehang","model":"m1"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(reqCtx)

	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rr.Code, rr.Body.String())
	}
	events := parseSSEEvents(rr.Body.String())
	metaCount := 0
	var opening map[string]any
	for _, ev := range events {
		if ev["type"] != "meta" {
			continue
		}
		metaCount++
		if m, ok := ev["meta"].(map[string]any); ok && m != nil && opening == nil {
			opening = m
		}
	}
	if metaCount != 1 {
		t.Fatalf("expected one meta event, got %d: %v", metaCount, events)
	}
	if opening == nil || opening["assistantMessageId"] == nil || opening["assistantMessageId"] == "" {
		t.Fatalf("opening meta: %+v", opening)
	}
	var cancelledPartial bool
	for _, ev := range events {
		if ev["type"] != "cancelled" {
			continue
		}
		if p, ok := ev["partial"].(bool); ok && p {
			cancelledPartial = true
			break
		}
	}
	if !cancelledPartial {
		t.Fatalf("expected cancelled with partial true, events=%v", events)
	}
}
