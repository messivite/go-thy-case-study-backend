package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	CreateChatSession(ctx context.Context, userID string, title string) (ChatSession, error)
	GetChatSessionsByUser(ctx context.Context, userID string) ([]ChatSession, error)
	GetChatSessionByID(ctx context.Context, sessionID string) (ChatSession, error)
	SaveMessage(ctx context.Context, sessionID string, userID string, role string, content string) (ChatMessage, error)
	GetMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error)
}
