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
	CloseFileLog()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "test.event") || !strings.Contains(string(b), `"level":"info"`) {
		t.Fatalf("unexpected file content: %s", b)
	}
}
