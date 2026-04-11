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
	GetChatSessionsByUserPage(ctx context.Context, userID string, limit int, cursor *SessionCursor) (SessionListPage, error)
	GetChatSessionByID(ctx context.Context, sessionID string) (ChatSession, error)
	SoftDeleteChatSession(ctx context.Context, sessionID string) error
	UpdateSessionLastLLM(ctx context.Context, sessionID, provider, model string) error
	SaveMessage(ctx context.Context, sessionID, userID string, role Role, content, provider, model string) (ChatMessage, error)
	SaveAssistantPlaceholder(ctx context.Context, sessionID, messageID, provider, model string) (ChatMessage, error)
	UpdateAssistantMessageContent(ctx context.Context, sessionID, messageID, content, provider, model string) (ChatMessage, error)
	SoftDeleteChatMessageByID(ctx context.Context, sessionID, messageID string) error
	SaveMessages(ctx context.Context, sessionID, userID string, messages []BatchMessage) ([]ChatMessage, error)
	SoftDeleteUserMessage(ctx context.Context, sessionID, messageID, userID string) error
	GetMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error)
	GetMessagesBySessionPage(ctx context.Context, sessionID string, limit int, direction string, cursor *MessageCursor) ([]ChatMessage, int, error)
	SearchChats(ctx context.Context, params SearchChatParams) (SearchChatsResult, error)
	GetUserProfile(ctx context.Context, userID string) (UserProfile, error)
	PatchUserProfile(ctx context.Context, userID string, patch ProfilePatch) (UserProfile, error)
	UploadUserAvatarJPEG(ctx context.Context, userID string, jpeg []byte) (publicURL string, err error)
}
