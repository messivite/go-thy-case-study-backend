package repo

import (
	"context"
	"strings"
	"testing"
	"time"

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

func TestMemoryRepositoryGetChatSessionsByUser_excludesSoftDeleted(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()
	const uid = "user-soft-del-list"

	keep, err := r.CreateChatSession(ctx, uid, "keep", "", "")
	if err != nil {
		t.Fatal(err)
	}
	gone, err := r.CreateChatSession(ctx, uid, "gone", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.SoftDeleteChatSession(ctx, gone.ID.String()); err != nil {
		t.Fatal(err)
	}

	list, err := r.GetChatSessionsByUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(list))
	}
	if list[0].ID != keep.ID {
		t.Fatalf("expected kept session %s, got %s", keep.ID, list[0].ID)
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

func TestMemoryRepositoryGetChatSessionsByUserPage_empty(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()
	page, err := r.GetChatSessionsByUserPage(ctx, "nobody", 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalCount != 0 || len(page.Items) != 0 {
		t.Fatalf("expected empty page, got total=%d items=%d", page.TotalCount, len(page.Items))
	}
}

func TestMemoryRepositoryGetChatSessionsByUserPage_paginationAndPreview(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()
	const uid = "user-page-1"

	older, err := r.CreateChatSession(ctx, uid, "older", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.SaveMessage(ctx, older.ID.String(), uid, domain.RoleUser, "msg-old", "", ""); err != nil {
		t.Fatal(err)
	}

	newer, err := r.CreateChatSession(ctx, uid, "newer", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.SaveMessage(ctx, newer.ID.String(), uid, domain.RoleUser, "msg-new", "", ""); err != nil {
		t.Fatal(err)
	}

	// Full list: newest session first by last message time
	pageAll, err := r.GetChatSessionsByUserPage(ctx, uid, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if pageAll.TotalCount != 2 || len(pageAll.Items) != 2 {
		t.Fatalf("total=%d items=%d", pageAll.TotalCount, len(pageAll.Items))
	}
	if pageAll.Items[0].Session.ID != newer.ID {
		t.Fatalf("expected newer session first, got %s", pageAll.Items[0].Session.ID)
	}
	if pageAll.Items[0].LastMessagePreview != "msg-new" {
		t.Fatalf("preview: %q", pageAll.Items[0].LastMessagePreview)
	}

	// First page: limit 1 => only newest
	p1, err := r.GetChatSessionsByUserPage(ctx, uid, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p1.Items) != 1 || p1.Items[0].Session.ID != newer.ID {
		t.Fatalf("p1: %+v", p1.Items)
	}

	cur := &domain.SessionCursor{SortAt: p1.Items[0].SortAt, SessionID: p1.Items[0].Session.ID.String()}
	p2, err := r.GetChatSessionsByUserPage(ctx, uid, 1, cur)
	if err != nil {
		t.Fatal(err)
	}
	if len(p2.Items) != 1 || p2.Items[0].Session.ID != older.ID {
		t.Fatalf("p2: %+v", p2.Items)
	}
}

func TestMemoryRepositoryGetChatSessionsByUserPage_longPreviewSkipsDeleted(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()
	const uid = "user-page-2"

	s, err := r.CreateChatSession(ctx, uid, "t", "", "")
	if err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("x", 90)
	if _, err := r.SaveMessage(ctx, s.ID.String(), uid, domain.RoleUser, long, "", ""); err != nil {
		t.Fatal(err)
	}
	page, err := r.GetChatSessionsByUserPage(ctx, uid, 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items=%d", len(page.Items))
	}
	if len(page.Items[0].LastMessagePreview) != 80 {
		t.Fatalf("expected 80 runes preview, got %d", len(page.Items[0].LastMessagePreview))
	}

	// Two messages: delete the newer (last); preview should fall back to older content
	s2, err := r.CreateChatSession(ctx, uid, "t2", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.SaveMessage(ctx, s2.ID.String(), uid, domain.RoleUser, "keep-me", "", ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	mNew, err := r.SaveMessage(ctx, s2.ID.String(), uid, domain.RoleUser, "delete-me", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.SoftDeleteUserMessage(ctx, s2.ID.String(), mNew.ID.String(), uid); err != nil {
		t.Fatal(err)
	}
	page2, err := r.GetChatSessionsByUserPage(ctx, uid, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, it := range page2.Items {
		if it.Session.ID == s2.ID && it.LastMessagePreview == "keep-me" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected preview keep-me after soft-delete last msg, items=%+v", page2.Items)
	}
}

func TestMemoryRepositorySupportedModels_syncListSortAndActive(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	if err := r.SyncSupportedModels(ctx, []domain.SupportedModel{
		{Provider: "z", ModelID: "m2", DisplayName: "Z2", SupportsStream: false},
		{Provider: "a", ModelID: "m1", DisplayName: "A1", SupportsStream: true},
	}); err != nil {
		t.Fatal(err)
	}

	list, err := r.ListActiveSupportedModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("list len: %d", len(list))
	}
	if list[0].Provider != "a" || list[0].ModelID != "m1" {
		t.Fatalf("expected first a/m1, got %s/%s", list[0].Provider, list[0].ModelID)
	}
	if list[1].Provider != "z" || list[1].ModelID != "m2" {
		t.Fatalf("expected second z/m2, got %s/%s", list[1].Provider, list[1].ModelID)
	}

	ok, err := r.IsModelActive(ctx, "A", "m1")
	if err != nil || !ok {
		t.Fatalf("IsModelActive A/m1: ok=%v err=%v", ok, err)
	}
	ok, err = r.IsModelActive(ctx, "a", "  m1  ")
	if err != nil || !ok {
		t.Fatalf("IsModelActive trim: ok=%v err=%v", ok, err)
	}
	ok, err = r.IsModelActive(ctx, "a", "missing")
	if err != nil || ok {
		t.Fatalf("IsModelActive missing: ok=%v err=%v", ok, err)
	}
	ok, err = r.IsModelActive(ctx, "", "m1")
	if err != nil || ok {
		t.Fatalf("IsModelActive empty provider: ok=%v err=%v", ok, err)
	}
	ok, err = r.IsModelActive(ctx, "a", "")
	if err != nil || ok {
		t.Fatalf("IsModelActive empty model: ok=%v err=%v", ok, err)
	}
}
