package chat

type NormalizedUsage struct {
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

func NormalizeUsage(raw map[string]any) NormalizedUsage {
	u := NormalizedUsage{
		Provider: getString(raw, "provider"),
		Model:    getString(raw, "model"),
	}
	u.PromptTokens = getInt(raw, "prompt_tokens")
	u.CompletionTokens = getInt(raw, "completion_tokens")
	u.TotalTokens = getInt(raw, "total_tokens")

	if u.TotalTokens == 0 && (u.PromptTokens > 0 || u.CompletionTokens > 0) {
		u.TotalTokens = u.PromptTokens + u.CompletionTokens
	}

	return u
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	default:
		return 0
	}
}
