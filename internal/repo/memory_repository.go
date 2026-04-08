package repo

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]ChatSession
	messages map[uuid.UUID][]ChatMessage
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		sessions: make(map[uuid.UUID]ChatSession),
		messages: make(map[uuid.UUID][]ChatMessage),
	}
}

func (r *MemoryRepository) CreateChatSession(ctx context.Context, userID string, title string) (ChatSession, error) {
	id := uuid.New()
	session := ChatSession{
		ID:        id,
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now().UTC(),
	}

	r.mu.Lock()
	r.sessions[id] = session
	r.mu.Unlock()

	return session, nil
}

func (r *MemoryRepository) GetChatSessionsByUser(ctx context.Context, userID string) ([]ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]ChatSession, 0)
	for _, session := range r.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}

	slices.SortFunc(sessions, func(a, b ChatSession) int {
		if a.CreatedAt.Equal(b.CreatedAt) {
			return 0
		}
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		return 1
	})

	return sessions, nil
}

func (r *MemoryRepository) GetChatSessionByID(ctx context.Context, sessionID string) (ChatSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return ChatSession{}, fmt.Errorf("invalid session id: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return ChatSession{}, fmt.Errorf("session not found")
	}
	return session, nil
}

func (r *MemoryRepository) SaveMessage(ctx context.Context, sessionID string, userID string, role string, content string) (ChatMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("invalid session id: %w", err)
	}

	message := ChatMessage{
		ID:        uuid.New(),
		SessionID: sessionUUID,
		UserID:    userID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[sessionUUID]; !ok {
		return ChatMessage{}, fmt.Errorf("session not found")
	}

	r.messages[sessionUUID] = append(r.messages[sessionUUID], message)
	return message, nil
}

func (r *MemoryRepository) GetMessagesBySession(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id: %w", err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.sessions[sessionUUID]; !ok {
		return nil, fmt.Errorf("session not found")
	}

	msgs := r.messages[sessionUUID]
	out := make([]ChatMessage, len(msgs))
	copy(out, msgs)
	return out, nil
}
