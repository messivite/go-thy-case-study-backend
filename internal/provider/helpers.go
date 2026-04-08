package provider

import (
	"strings"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

func lastUserContent(messages []domain.ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == domain.RoleUser && strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}
	return ""
}

func splitChunks(s string, n int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{""}
	}
	if n <= 0 || len(s) <= n {
		return []string{s}
	}

	out := make([]string, 0, len(s)/n+1)
	for i := 0; i < len(s); i += n {
		j := i + n
		if j > len(s) {
			j = len(s)
		}
		out = append(out, s[i:j])
	}
	return out
}
