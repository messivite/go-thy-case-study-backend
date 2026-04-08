package chat

import "context"

type Repository interface {
	CreateChatSession(ctx context.Context, userID, title string) (ChatSession, error)
	GetChatSessionsByUser(ctx context.Context, userID string) ([]ChatSession, error)
	GetChatSessionByID(ctx context.Context, sessionID string) (ChatSession, error)
	SaveMessage(ctx context.Context, sessionID, userID string, role Role, content string) (ChatMessage, error)
	GetMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error)
}
