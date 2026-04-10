package chat

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

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

