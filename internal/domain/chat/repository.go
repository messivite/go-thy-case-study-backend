package chat

import "context"

type BatchMessage struct {
	Content  string
	Provider string
	Model    string
}

type Repository interface {
	CreateChatSession(ctx context.Context, userID, title, defaultProvider, defaultModel string) (ChatSession, error)
	GetChatSessionsByUser(ctx context.Context, userID string) ([]ChatSession, error)
	GetChatSessionByID(ctx context.Context, sessionID string) (ChatSession, error)
	UpdateSessionLastLLM(ctx context.Context, sessionID, provider, model string) error
	SaveMessage(ctx context.Context, sessionID, userID string, role Role, content, provider, model string) (ChatMessage, error)
	SaveMessages(ctx context.Context, sessionID, userID string, messages []BatchMessage) ([]ChatMessage, error)
	GetMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error)
}
