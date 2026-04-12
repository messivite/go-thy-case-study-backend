package chat

import (
	"context"
	"errors"
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

type streamFailProvider struct{}

func (streamFailProvider) Name() string { return "streamfail" }

func (streamFailProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_, _ = ctx, req
	return domain.ProviderResponse{}, errors.New("no complete")
}

func (streamFailProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_, _ = ctx, req
	return nil, errors.New("stream start failed")
}

type streamImmediateCloseProvider struct{}

func (streamImmediateCloseProvider) Name() string { return "streamempty" }

func (streamImmediateCloseProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_, _ = ctx, req
	return domain.ProviderResponse{Content: ""}, nil
}

func (streamImmediateCloseProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_, _ = ctx, req
	ch := make(chan domain.StreamEvent)
	close(ch)
	return ch, nil
}

// streamBlockProvider keeps the stream open until ctx is cancelled (simulates in-flight SSE before stop).
type streamBlockProvider struct{}

func (streamBlockProvider) Name() string { return "streamblock" }

func (streamBlockProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	_, _ = ctx, req
	return domain.ProviderResponse{Content: ""}, nil
}

func (streamBlockProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	_, _ = ctx, req
	ch := make(chan domain.StreamEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

func TestUseCase_StreamMessage_usageMetaIncludesUserAndAssistantMessageIDs(t *testing.T) {
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

	events, usage, finalize, _, err := uc.StreamMessage(ctx, "u1", s.ID.String(), "streamtest", "m1", "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	uid, ok := usage["userMessageId"].(string)
	if !ok || uid == "" {
		t.Fatalf("expected userMessageId in usage meta, got %+v", usage)
	}
	aid, ok := usage["assistantMessageId"].(string)
	if !ok || aid == "" {
		t.Fatalf("expected assistantMessageId in usage meta, got %+v", usage)
	}

	var b strings.Builder
	for ev := range events {
		if ev.Type == domain.EventDelta {
			b.WriteString(ev.Delta)
		}
	}
	msg, err := finalize(b.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ID.String() != aid {
		t.Fatalf("finalize id %s want placeholder %s", msg.ID.String(), aid)
	}
	if msg.Role != domain.RoleAssistant || msg.Content != "ok" {
		t.Fatalf("assistant: %+v", msg)
	}
}

func TestUseCase_StreamMessage_removesPlaceholderWhenStreamFails(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "streamfail", ModelID: "m1", DisplayName: "M1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	s, err := mem.CreateChatSession(ctx, "u1", "t", "streamfail", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("streamfail")
	reg.Register(streamFailProvider{}, provider.ProviderMeta{
		Name: "streamfail", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	_, _, _, _, err = uc.StreamMessage(ctx, "u1", s.ID.String(), "streamfail", "m1", "hello", nil)
	if err == nil {
		t.Fatal("expected stream error")
	}
	msgs, err := mem.GetMessagesBySession(ctx, s.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Role != domain.RoleUser {
		t.Fatalf("expected only user message after rollback, got %d msgs %+v", len(msgs), msgs)
	}
}

func TestUseCase_StreamMessage_finalizeEmpty_softDeletesPlaceholder(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "streamempty", ModelID: "m1", DisplayName: "M1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	s, err := mem.CreateChatSession(ctx, "u1", "t", "streamempty", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("streamempty")
	reg.Register(streamImmediateCloseProvider{}, provider.ProviderMeta{
		Name: "streamempty", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	events, _, finalize, _, err := uc.StreamMessage(ctx, "u1", s.ID.String(), "streamempty", "m1", "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	for range events {
	}
	_, err = finalize("", nil)
	if err == nil {
		t.Fatal("expected empty finalize error")
	}
	if !errors.Is(err, domain.ErrMissingContent) {
		t.Fatalf("expected ErrMissingContent, got %v", err)
	}
	msgs, err := mem.GetMessagesBySession(ctx, s.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Role != domain.RoleUser {
		t.Fatalf("expected only user after empty finalize, got %d %+v", len(msgs), msgs)
	}
}

func TestUseCase_StreamMessage_cancelZeroRollsBackUserMessage(t *testing.T) {
	ctx := context.Background()
	mem := repo.NewMemoryRepository()
	if err := mem.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "streamblock", ModelID: "m1", DisplayName: "M1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}
	s, err := mem.CreateChatSession(ctx, "u1", "t", "streamblock", "m1")
	if err != nil {
		t.Fatal(err)
	}
	reg := provider.NewRegistry("streamblock")
	reg.Register(streamBlockProvider{}, provider.ProviderMeta{
		Name: "streamblock", DefaultModel: "m1", RequiredEnvKey: "X", SupportsStream: true,
	})
	uc := NewUseCase(mem, &meUsageQuotaStub{q: domain.UserQuota{QuotaBypass: true}}, reg, mem)

	before, err := mem.GetMessagesBySession(ctx, s.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	n0 := len(before)

	streamCtx, streamStop := context.WithCancel(ctx)
	events, _, _, cancel, err := uc.StreamMessage(streamCtx, "u1", s.ID.String(), "streamblock", "m1", "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	cancel(0)
	streamStop()
	for range events {
	}

	after, err := mem.GetMessagesBySession(ctx, s.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != n0 {
		t.Fatalf("stop with no partial should roll back user+assistant; got %d messages want %d", len(after), n0)
	}
}
