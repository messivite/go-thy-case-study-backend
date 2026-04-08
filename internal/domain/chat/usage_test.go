package chat

import "testing"

func TestNormalizeUsage(t *testing.T) {
	raw := map[string]any{
		"provider":          "openai",
		"model":             "gpt-4.1-mini",
		"prompt_tokens":     14,
		"completion_tokens": 10,
		"total_tokens":      24,
	}

	u := NormalizeUsage(raw)

	if u.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", u.Provider)
	}
	if u.Model != "gpt-4.1-mini" {
		t.Errorf("expected model 'gpt-4.1-mini', got %q", u.Model)
	}
	if u.PromptTokens != 14 {
		t.Errorf("expected prompt_tokens 14, got %d", u.PromptTokens)
	}
	if u.CompletionTokens != 10 {
		t.Errorf("expected completion_tokens 10, got %d", u.CompletionTokens)
	}
	if u.TotalTokens != 24 {
		t.Errorf("expected total_tokens 24, got %d", u.TotalTokens)
	}
}

func TestNormalizeUsageFloat64(t *testing.T) {
	raw := map[string]any{
		"provider":          "gemini",
		"model":             "gemini-2.0-flash",
		"prompt_tokens":     float64(100),
		"completion_tokens": float64(50),
		"total_tokens":      float64(150),
	}

	u := NormalizeUsage(raw)
	if u.PromptTokens != 100 {
		t.Errorf("expected 100, got %d", u.PromptTokens)
	}
	if u.TotalTokens != 150 {
		t.Errorf("expected 150, got %d", u.TotalTokens)
	}
}

func TestNormalizeUsageAutoTotal(t *testing.T) {
	raw := map[string]any{
		"prompt_tokens":     20,
		"completion_tokens": 30,
	}

	u := NormalizeUsage(raw)
	if u.TotalTokens != 50 {
		t.Errorf("expected auto-calculated total 50, got %d", u.TotalTokens)
	}
}

func TestNormalizeUsageEmpty(t *testing.T) {
	u := NormalizeUsage(map[string]any{})
	if u.Provider != "" || u.Model != "" || u.TotalTokens != 0 {
		t.Errorf("expected empty usage, got %+v", u)
	}
}
