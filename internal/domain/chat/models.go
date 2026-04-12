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
	ID              uuid.UUID
	UserID          string
	Title           string
	CreatedAt       time.Time
	DeletedAt       *time.Time
	LastProvider    string
	LastModel       string
	DefaultProvider string
	DefaultModel    string
}

type ChatMessage struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	UserID    string
	Role      Role
	Content   string
	CreatedAt time.Time
	DeletedAt *time.Time
	Provider  string
	Model     string
	// Liked: nil = log yok; true = action 1 (liked); false = action 2 (unlike) satırı var. System/silinmiş mesajda nil.
	Liked *bool
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

// MessageLikeSyncItem: offline kuyruktan senkron; action 1=like, 2=unlike.
type MessageLikeSyncItem struct {
	MessageID string
	Action    int
}

// MessageLikeSyncResult: satır bazlı sonuç (ok=false iken Code dolu).
type MessageLikeSyncResult struct {
	MessageID string
	OK        bool
	State     int    // 1 liked, 2 unliked when OK
	Code      string // e.g. message_not_found, message_not_likeable
}
