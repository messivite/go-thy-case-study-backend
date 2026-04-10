package chat

import (
	"testing"
	"time"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

func TestSearchCursorEncodeDecode_RoundTrip(t *testing.T) {
	in := domain.SearchCursor{
		SortAt:    time.Date(2026, 4, 10, 12, 30, 0, 123, time.UTC),
		SessionID: "a3a5e3cb-4f50-4e48-9107-a2a985a7134e",
	}
	token := encodeSearchCursor(in)
	out, err := decodeSearchCursor(token)
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("decoded cursor is nil")
	}
	if !out.SortAt.Equal(in.SortAt) || out.SessionID != in.SessionID {
		t.Fatalf("cursor mismatch: in=%+v out=%+v", in, *out)
	}
}

func TestDecodeSearchCursor_Invalid(t *testing.T) {
	if _, err := decodeSearchCursor("not-base64"); err == nil {
		t.Fatal("expected invalid cursor error")
	}
}
