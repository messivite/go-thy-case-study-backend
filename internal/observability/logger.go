package observability

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp string         `json:"ts"`
	Level     string         `json:"level"`
	Event     string         `json:"event"`
	Fields    map[string]any `json:"fields,omitempty"`
	Error     string         `json:"error,omitempty"`
}

var sinkMu sync.Mutex
var sinkFile *os.File

// EnableFileLog appends her yapılandırılmış log satırını JSON Lines olarak path dosyasına yazar.
// Boş path: dosya sink’i kapatır. Process çıkarken CloseFileLog çağır.
func EnableFileLog(path string) error {
	sinkMu.Lock()
	defer sinkMu.Unlock()
	if sinkFile != nil {
		_ = sinkFile.Close()
		sinkFile = nil
	}
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	sinkFile = f
	return nil
}

// CloseFileLog dosya sink’ini kapatır (graceful shutdown).
func CloseFileLog() {
	sinkMu.Lock()
	defer sinkMu.Unlock()
	if sinkFile != nil {
		_ = sinkFile.Close()
		sinkFile = nil
	}
}

func Info(event string, fields map[string]any) {
	emit("info", event, fields, "")
}

func Warn(event string, fields map[string]any) {
	emit("warn", event, fields, "")
}

func Error(event string, err error, fields map[string]any) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	emit("error", event, fields, errMsg)
}

func LLMRequest(provider, model, userID, sessionID string) {
	Info("llm.request", map[string]any{
		"provider":   provider,
		"model":      model,
		"user_id":    userID,
		"session_id": sessionID,
	})
}

func LLMResponse(provider, model string, promptTokens, completionTokens, totalTokens int, latencyMs int64) {
	Info("llm.response", map[string]any{
		"provider":          provider,
		"model":             model,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
		"latency_ms":        latencyMs,
	})
}

func LLMError(provider, model string, err error) {
	Error("llm.error", err, map[string]any{
		"provider": provider,
		"model":    model,
	})
}

func LLMCancelled(provider, model, userID, sessionID string, partialChars int) {
	Info("llm.cancelled", map[string]any{
		"provider":      provider,
		"model":         model,
		"user_id":       userID,
		"session_id":    sessionID,
		"partial_chars": partialChars,
	})
}

func emit(level, event string, fields map[string]any, errMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Event:     event,
		Fields:    fields,
		Error:     errMsg,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("observability: marshal: %v", err)
		return
	}
	line := append(data, '\n')
	log.Print(string(data))

	sinkMu.Lock()
	if sinkFile != nil {
		_, _ = sinkFile.Write(line)
	}
	sinkMu.Unlock()
}
