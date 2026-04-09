package repo

import (
	"context"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

func TestMemoryRepositorySessionLifecycle(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	session, err := r.CreateChatSession(ctx, "user-1", "hello", "", "")
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

	if err := r.UpdateSessionLastLLM(ctx, session.ID.String(), "gemini", "gemini-2.5-flash"); err != nil {
		t.Fatalf("update last llm: %v", err)
	}
	got, err = r.GetChatSessionByID(ctx, session.ID.String())
	if err != nil {
		t.Fatalf("get session after update: %v", err)
	}
	if got.LastProvider != "gemini" || got.LastModel != "gemini-2.5-flash" {
		t.Fatalf("last llm: got provider=%q model=%q", got.LastProvider, got.LastModel)
	}
}

func TestMemoryRepositoryMessages(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	session, err := r.CreateChatSession(ctx, "user-2", "chat", "", "")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	_, err = r.SaveMessage(ctx, session.ID.String(), "user-2", domain.RoleUser, "hi", "", "")
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

func TestMemoryRepositorySaveMessages(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	session, err := r.CreateChatSession(ctx, "user-3", "sync", "", "")
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	saved, err := r.SaveMessages(ctx, session.ID.String(), "user-3", []domain.BatchMessage{
		{Content: "offline 1", Provider: "openai", Model: "gpt-4o"},
		{Content: "offline 2", Provider: "openai", Model: "gpt-4o"},
	})
	if err != nil {
		t.Fatalf("save messages failed: %v", err)
	}
	if len(saved) != 2 {
		t.Fatalf("expected 2 saved messages, got %d", len(saved))
	}
	if saved[0].Role != domain.RoleUser || saved[1].Role != domain.RoleUser {
		t.Fatalf("expected user role for all saved messages")
	}
	if saved[0].Provider != "openai" || saved[1].Model != "gpt-4o" {
		t.Fatalf("provider/model mapping failed: %+v", saved)
	}

	all, err := r.GetMessagesBySession(ctx, session.ID.String())
	if err != nil {
		t.Fatalf("get messages failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 messages in session, got %d", len(all))
	}
	if all[0].Content != "offline 1" || all[1].Content != "offline 2" {
		t.Fatalf("unexpected message order/content: %+v", all)
	}
	if all[1].CreatedAt.Before(all[0].CreatedAt) {
		t.Fatalf("expected monotonic createdAt ordering")
	}
}

func TestMemoryRepositorySaveMessagesErrors(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	_, err := r.SaveMessages(ctx, "not-a-uuid", "user-1", []domain.BatchMessage{{Content: "x"}})
	if err == nil {
		t.Fatal("expected invalid session id error")
	}

	_, err = r.SaveMessages(ctx, "11111111-1111-1111-1111-111111111111", "user-1", []domain.BatchMessage{{Content: "x"}})
	if err == nil {
		t.Fatal("expected session not found error")
	}
}
