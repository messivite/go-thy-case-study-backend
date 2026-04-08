package repo

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

var _ domain.Repository = (*MemoryRepository)(nil)

type MemoryRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]domain.ChatSession
	messages map[uuid.UUID][]domain.ChatMessage
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		sessions: make(map[uuid.UUID]domain.ChatSession),
		messages: make(map[uuid.UUID][]domain.ChatMessage),
	}
}

func (r *MemoryRepository) CreateChatSession(ctx context.Context, userID, title string) (domain.ChatSession, error) {
	id := uuid.New()
	session := domain.ChatSession{
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

func (r *MemoryRepository) GetChatSessionsByUser(ctx context.Context, userID string) ([]domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]domain.ChatSession, 0)
	for _, session := range r.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}

	slices.SortFunc(sessions, func(a, b domain.ChatSession) int {
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

func (r *MemoryRepository) GetChatSessionByID(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.ChatSession{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return domain.ChatSession{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *MemoryRepository) SaveMessage(ctx context.Context, sessionID, userID string, role domain.Role, content string) (domain.ChatMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.ChatMessage{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	message := domain.ChatMessage{
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
		return domain.ChatMessage{}, domain.ErrSessionNotFound
	}

	r.messages[sessionUUID] = append(r.messages[sessionUUID], message)
	return message, nil
}

func (r *MemoryRepository) GetMessagesBySession(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.sessions[sessionUUID]; !ok {
		return nil, domain.ErrSessionNotFound
	}

	msgs := r.messages[sessionUUID]
	out := make([]domain.ChatMessage, len(msgs))
	copy(out, msgs)
	return out, nil
}
