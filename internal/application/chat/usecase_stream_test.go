package chat

import (
	"context"
	"strings"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

// streamTestProvider emits one delta then leaves the channel open until ctx is cancelled.
type streamTestProvider struct{}

func (streamTestProvider) Name() string { return "streamtest" }

func (streamTestProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_ = ctx
	_ = req
	return domain.ProviderResponse{Content: "ok"}, nil
}

func (streamTestProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_ = ctx
	_ = req
	ch := make(chan domain.StreamEvent, 1)
	ch <- domain.StreamEvent{Type: domain.EventDelta, Delta: "ok"}
	close(ch)
	return ch, nil
}

func TestUseCase_StreamMessage_usageMetaIncludesUserMessageID(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	s, err := mem.CreateChatSession(ctx, "u1", "t", "streamtest", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("streamtest")
	reg.Register(streamTestProvider{}, provider.ProviderMeta{
		Name: "streamtest", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	stub := &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}
	uc := NewUseCase(mem, stub, reg, nil)

	events, usage, finalize, cancel, err := uc.StreamMessage(ctx, "u1", s.ID.String(), "streamtest", "m1", "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	uid, ok := usage["userMessageId"].(string)
	if !ok || uid == "" {
		t.Fatalf("expected userMessageId in usage meta, got %+v", usage)
	}

	var b strings.Builder
	for ev := range events {
		if ev.Type == domain.EventDelta {
			b.WriteString(ev.Delta)
		}
	}
	msg, err := finalize(b.String())
	if err != nil {
		t.Fatal(err)
	}
	cancel(0)
	if msg.Role != domain.RoleAssistant || msg.Content != "ok" {
		t.Fatalf("assistant: %+v", msg)
	}
}
