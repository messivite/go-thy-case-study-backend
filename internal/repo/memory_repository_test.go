package repo

import (
	"context"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

func TestMemoryRepositorySessionLifecycle(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	session, err := r.CreateChatSession(ctx, "user-1", "hello")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	got, err := r.GetChatSessionByID(ctx, session.ID.String())
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}
	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", got.UserID)
	}
}

func TestMemoryRepositoryMessages(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	session, err := r.CreateChatSession(ctx, "user-2", "chat")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	_, err = r.SaveMessage(ctx, session.ID.String(), "user-2", domain.RoleUser, "hi")
	if err != nil {
		t.Fatalf("save message failed: %v", err)
	}

	messages, err := r.GetMessagesBySession(ctx, session.ID.String())
	if err != nil {
		t.Fatalf("get messages failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Content != "hi" {
		t.Fatalf("expected content hi, got %s", messages[0].Content)
	}
}
