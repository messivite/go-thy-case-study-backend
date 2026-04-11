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

func TestHandler_StreamMessage_initialMetaAndDoneIncludeUserMessageID(t *testing.T) {
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
	var firstMeta, doneMeta map[string]any
	for _, ev := range events {
		if ev["type"] != "meta" {
			continue
		}
		meta, _ := ev["meta"].(map[string]any)
		if meta == nil {
			continue
		}
		if _, ok := meta["userMessageId"]; ok && firstMeta == nil {
			firstMeta = meta
		}
		if _, ok := meta["assistantMessageId"]; ok {
			doneMeta = meta
		}
	}
	if firstMeta == nil {
		t.Fatalf("no opening meta with userMessageId: %v", events)
	}
	if firstMeta["userMessageId"] == nil || firstMeta["userMessageId"] == "" {
		t.Fatalf("first meta: %+v", firstMeta)
	}
	if doneMeta == nil {
		t.Fatalf("no done meta with assistantMessageId: %v", events)
	}
	if doneMeta["userMessageId"] == nil || doneMeta["userMessageId"] == "" {
		t.Fatalf("done meta missing userMessageId: %+v", doneMeta)
	}
	if doneMeta["assistantMessageId"] == nil || doneMeta["assistantMessageId"] == "" {
		t.Fatalf("done meta: %+v", doneMeta)
	}
}

func TestHandler_StreamMessage_cancelPartialMetaIncludesUserMessageID(t *testing.T) {
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
	var partialMeta map[string]any
	for _, ev := range events {
		if ev["type"] != "meta" {
			continue
		}
		meta, _ := ev["meta"].(map[string]any)
		if meta == nil {
			continue
		}
		if v, ok := meta["partial"]; ok && v == true {
			partialMeta = meta
			break
		}
	}
	if partialMeta == nil {
		t.Fatalf("expected partial meta, events=%v", events)
	}
	if partialMeta["userMessageId"] == nil || partialMeta["userMessageId"] == "" {
		t.Fatalf("partial meta: %+v", partialMeta)
	}
	if partialMeta["assistantMessageId"] == nil || partialMeta["assistantMessageId"] == "" {
		t.Fatalf("partial meta: %+v", partialMeta)
	}
}
