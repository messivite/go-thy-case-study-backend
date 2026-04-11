package repo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

func TestSupabaseRepository_SaveAssistantPlaceholder_invalidSessionID(t *testing.T) {
	r := NewSupabaseRepository("http://example.com", "k")
	_, err := r.SaveAssistantPlaceholder(context.Background(), "x", uuid.New().String(), "p", "m")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSupabaseRepository_SaveAssistantPlaceholder_invalidMessageID(t *testing.T) {
	r := NewSupabaseRepository("http://example.com", "k")
	_, err := r.SaveAssistantPlaceholder(context.Background(), uuid.New().String(), "y", "p", "m")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSupabaseRepository_UpdateAssistantMessageContent_invalidIDs(t *testing.T) {
	r := NewSupabaseRepository("http://example.com", "k")
	_, err := r.UpdateAssistantMessageContent(context.Background(), "bad", uuid.New().String(), "c", "p", "m")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = r.UpdateAssistantMessageContent(context.Background(), uuid.New().String(), "bad", "c", "p", "m")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSupabaseRepository_SoftDeleteChatMessageByID_invalidIDs(t *testing.T) {
	r := NewSupabaseRepository("http://example.com", "k")
	err := r.SoftDeleteChatMessageByID(context.Background(), "bad", uuid.New().String())
	if err == nil {
		t.Fatal("expected error")
	}
	err = r.SoftDeleteChatMessageByID(context.Background(), uuid.New().String(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSupabaseRepository_streamingMessageRPCs_viaREST(t *testing.T) {
	sid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mid := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	created := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/v1/chat_messages") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			var row map[string]any
			if json.Unmarshal(body, &row) != nil {
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			if row["id"] != mid.String() || row["content"] != "" || row["role"] != "assistant" {
				http.Error(w, "unexpected body", http.StatusBadRequest)
				return
			}
			resp := []map[string]any{{
				"id": mid.String(), "session_id": sid.String(), "role": "assistant",
				"content": "", "created_at": created, "provider": "openai", "model": "gpt-4o",
			}}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			var patch map[string]any
			if json.Unmarshal(body, &patch) != nil {
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			if r.URL.Query().Get("id") != "eq."+mid.String() {
				http.Error(w, "bad query", http.StatusBadRequest)
				return
			}
			_, softDelete := patch["deleted_at"]
			if softDelete && patch["content"] == nil {
				resp := []map[string]any{{
					"id": mid.String(), "session_id": sid.String(), "role": "assistant",
					"content": "x", "created_at": created, "deleted_at": patch["deleted_at"],
					"provider": "openai", "model": "gpt-4o",
				}}
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			if patch["content"] != "hello" {
				http.Error(w, "unexpected patch", http.StatusBadRequest)
				return
			}
			resp := []map[string]any{{
				"id": mid.String(), "session_id": sid.String(), "role": "assistant",
				"content": "hello", "created_at": created, "provider": "openai", "model": "gpt-4o",
			}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.Error(w, "method", http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	repo := NewSupabaseRepository(ts.URL, "service-key")
	ctx := context.Background()

	ph, err := repo.SaveAssistantPlaceholder(ctx, sid.String(), mid.String(), "openai", "gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if ph.ID != mid || ph.Content != "" {
		t.Fatalf("placeholder %+v", ph)
	}

	upd, err := repo.UpdateAssistantMessageContent(ctx, sid.String(), mid.String(), "hello", "openai", "gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if upd.Content != "hello" {
		t.Fatalf("update %+v", upd)
	}

	if err := repo.SoftDeleteChatMessageByID(ctx, sid.String(), mid.String()); err != nil {
		t.Fatal(err)
	}
}

func TestSupabaseRepository_UpdateAssistantMessageContent_emptyRows(t *testing.T) {
	sid := uuid.New().String()
	mid := uuid.New().String()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer ts.Close()
	repo := NewSupabaseRepository(ts.URL, "k")
	_, err := repo.UpdateAssistantMessageContent(context.Background(), sid, mid, "c", "", "")
	if err != domain.ErrMessageNotFound {
		t.Fatalf("got %v want ErrMessageNotFound", err)
	}
}

func TestSupabaseRepository_SoftDeleteChatMessageByID_emptyRows(t *testing.T) {
	sid := uuid.New().String()
	mid := uuid.New().String()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer ts.Close()
	repo := NewSupabaseRepository(ts.URL, "k")
	err := repo.SoftDeleteChatMessageByID(context.Background(), sid, mid)
	if err != domain.ErrMessageNotFound {
		t.Fatalf("got %v want ErrMessageNotFound", err)
	}
}

func TestSupabaseRepository_SaveAssistantPlaceholder_noRowsReturned(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer ts.Close()
	repo := NewSupabaseRepository(ts.URL, "k")
	_, err := repo.SaveAssistantPlaceholder(context.Background(), uuid.New().String(), uuid.New().String(), "", "")
	if err == nil || !strings.Contains(err.Error(), "no rows returned") {
		t.Fatalf("expected no rows error, got %v", err)
	}
}
