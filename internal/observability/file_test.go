package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnableFileLogWritesJSONL(t *testing.T) {
	defer CloseFileLog()

	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := EnableFileLog(path); err != nil {
		t.Fatal(err)
	}

	Info("test.event", map[string]any{"k": 1})
	LLMCancelled("openai", "gpt-4o", "u1", "s1", 12)
	CloseFileLog()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, "test.event") || !strings.Contains(content, `"level":"info"`) {
		t.Fatalf("unexpected file content: %s", b)
	}
	if !strings.Contains(content, "llm.cancelled") || !strings.Contains(content, `"partial_chars":12`) {
		t.Fatalf("missing cancelled log content: %s", content)
	}
}
