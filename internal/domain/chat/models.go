package chat

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	}
	return false
}

type ChatSession struct {
	ID           uuid.UUID
	UserID       string
	Title        string
	CreatedAt    time.Time
	LastProvider string
	LastModel    string
}

type ChatMessage struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	UserID    string
	Role      Role
	Content   string
	CreatedAt time.Time
}

type StreamEventType string

const (
	EventDelta StreamEventType = "delta"
	EventDone  StreamEventType = "done"
	EventError StreamEventType = "error"
	EventMeta  StreamEventType = "meta"
)

type StreamEvent struct {
	Type    StreamEventType `json:"type"`
	Delta   string          `json:"delta,omitempty"`
	Message string          `json:"message,omitempty"`
	Meta    any             `json:"meta,omitempty"`
}

type ProviderRequest struct {
	Provider string
	Model    string
	Messages []ChatMessage
}

type ProviderResponse struct {
	Content string
	Usage   map[string]any
}
