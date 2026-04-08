package observability

import (
	"encoding/json"
	"log"
	"time"
)

type LogEntry struct {
	Timestamp string         `json:"ts"`
	Level     string         `json:"level"`
	Event     string         `json:"event"`
	Fields    map[string]any `json:"fields,omitempty"`
	Error     string         `json:"error,omitempty"`
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

func emit(level, event string, fields map[string]any, errMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Event:     event,
		Fields:    fields,
		Error:     errMsg,
	}
	data, _ := json.Marshal(entry)
	log.Println(string(data))
}
