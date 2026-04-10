package provider

import (
	"testing"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

func TestToAnthropicMessages_systemAndCollapse(t *testing.T) {
	msgs := []domain.ChatMessage{
		{Role: domain.RoleSystem, Content: "You are helpful."},
		{Role: domain.RoleUser, Content: "a"},
		{Role: domain.RoleUser, Content: "b"},
		{Role: domain.RoleAssistant, Content: "ok"},
		{Role: domain.RoleUser, Content: "c"},
	}
	sys, out := toAnthropicMessages(msgs)
	if sys != "You are helpful." {
		t.Fatalf("system: %q", sys)
	}
	if len(out) != 3 {
		t.Fatalf("len=%d %+v", len(out), out)
	}
	if out[0].Role != "user" || out[0].Content != "a\n\nb" {
		t.Fatalf("merged user: %+v", out[0])
	}
	if out[1].Role != "assistant" || out[1].Content != "ok" {
		t.Fatalf("assistant: %+v", out[1])
	}
	if out[2].Role != "user" || out[2].Content != "c" {
		t.Fatalf("last user: %+v", out[2])
	}
}
