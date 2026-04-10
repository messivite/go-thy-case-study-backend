package repo

import (
	"context"
	"fmt"
	"slices"
	"strings"
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

func (r *MemoryRepository) CreateChatSession(ctx context.Context, userID, title, defaultProvider, defaultModel string) (domain.ChatSession, error) {
	id := uuid.New()
	session := domain.ChatSession{
		ID:              id,
		UserID:          userID,
		Title:           title,
		CreatedAt:       time.Now().UTC(),
		DefaultProvider: defaultProvider,
		DefaultModel:    defaultModel,
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
		if session.UserID == userID && session.DeletedAt == nil {
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

func (r *MemoryRepository) GetChatSessionsByUserPage(ctx context.Context, userID string, limit int, cursor *domain.SessionCursor) (domain.SessionListPage, error) {
	sessions, err := r.GetChatSessionsByUser(ctx, userID)
	if err != nil {
		return domain.SessionListPage{}, err
	}
	items := make([]domain.SessionListItem, 0, len(sessions))
	for _, s := range sessions {
		msgs := r.messages[s.ID]
		lastPreview := ""
		updatedAt := s.CreatedAt
		if len(msgs) > 0 {
			last := msgs[len(msgs)-1]
			updatedAt = last.CreatedAt
			lastPreview = last.Content
			if len(lastPreview) > 80 {
				lastPreview = lastPreview[:80]
			}
		}
		sortAt := updatedAt
		items = append(items, domain.SessionListItem{
			Session:            s,
			LastMessagePreview: lastPreview,
			UpdatedAt:          updatedAt,
			SortAt:             sortAt,
		})
	}
	slices.SortFunc(items, func(a, b domain.SessionListItem) int {
		if a.SortAt.After(b.SortAt) {
			return -1
		}
		if a.SortAt.Before(b.SortAt) {
			return 1
		}
		if a.Session.ID.String() > b.Session.ID.String() {
			return -1
		}
		if a.Session.ID.String() < b.Session.ID.String() {
			return 1
		}
		return 0
	})
	total := len(items)
	filtered := items
	if cursor != nil {
		filtered = make([]domain.SessionListItem, 0, len(items))
		for _, it := range items {
			id := it.Session.ID.String()
			if it.SortAt.Before(cursor.SortAt) || (it.SortAt.Equal(cursor.SortAt) && id < cursor.SessionID) {
				filtered = append(filtered, it)
			}
		}
	}
	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	return domain.SessionListPage{
		TotalCount: total,
		Items:      filtered[:limit],
	}, nil
}

func (r *MemoryRepository) GetChatSessionByID(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.ChatSession{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok || session.DeletedAt != nil {
		return domain.ChatSession{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *MemoryRepository) SoftDeleteChatSession(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[id]
	if !ok || s.DeletedAt != nil {
		return domain.ErrSessionNotFound
	}
	now := time.Now().UTC()
	s.DeletedAt = &now
	r.sessions[id] = s
	return nil
}

func (r *MemoryRepository) UpdateSessionLastLLM(ctx context.Context, sessionID, provider, model string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[id]
	if !ok {
		return domain.ErrSessionNotFound
	}
	s.LastProvider = provider
	s.LastModel = model
	r.sessions[id] = s
	return nil
}

func (r *MemoryRepository) SaveMessage(ctx context.Context, sessionID, userID string, role domain.Role, content, provider, model string) (domain.ChatMessage, error) {
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
		Provider:  provider,
		Model:     model,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[sessionUUID]; !ok {
		return domain.ChatMessage{}, domain.ErrSessionNotFound
	}
	if r.sessions[sessionUUID].DeletedAt != nil {
		return domain.ChatMessage{}, domain.ErrSessionNotFound
	}

	r.messages[sessionUUID] = append(r.messages[sessionUUID], message)
	return message, nil
}

func (r *MemoryRepository) SaveMessages(ctx context.Context, sessionID, userID string, messages []domain.BatchMessage) ([]domain.ChatMessage, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[sessionUUID]; !ok {
		return nil, domain.ErrSessionNotFound
	}
	if r.sessions[sessionUUID].DeletedAt != nil {
		return nil, domain.ErrSessionNotFound
	}

	saved := make([]domain.ChatMessage, 0, len(messages))
	now := time.Now().UTC()
	for i, msg := range messages {
		createdAt := now.Add(time.Duration(i) * time.Millisecond)
		savedMsg := domain.ChatMessage{
			ID:        uuid.New(),
			SessionID: sessionUUID,
			UserID:    userID,
			Role:      domain.RoleUser,
			Content:   msg.Content,
			CreatedAt: createdAt,
			Provider:  msg.Provider,
			Model:     msg.Model,
		}
		r.messages[sessionUUID] = append(r.messages[sessionUUID], savedMsg)
		saved = append(saved, savedMsg)
	}

	return saved, nil
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
	if r.sessions[sessionUUID].DeletedAt != nil {
		return nil, domain.ErrSessionNotFound
	}

	msgs := r.messages[sessionUUID]
	out := make([]domain.ChatMessage, len(msgs))
	copy(out, msgs)
	return out, nil
}

func (r *MemoryRepository) GetMessagesBySessionPage(ctx context.Context, sessionID string, limit int, direction string, cursor *domain.MessageCursor) ([]domain.ChatMessage, int, error) {
	msgs, err := r.GetMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, 0, err
	}
	if direction == "" {
		direction = "older"
	}
	if direction != "older" && direction != "newer" {
		return nil, 0, domain.ErrInvalidDirection
	}
	total := len(msgs)
	if limit <= 0 || limit > total {
		limit = total
	}

	ordered := make([]domain.ChatMessage, 0, total)
	if direction == "older" {
		for i := len(msgs) - 1; i >= 0; i-- {
			ordered = append(ordered, msgs[i]) // desc
		}
	} else {
		ordered = append(ordered, msgs...) // asc
	}
	filtered := ordered
	if cursor != nil {
		filtered = make([]domain.ChatMessage, 0, len(ordered))
		for _, m := range ordered {
			id := m.ID.String()
			if direction == "older" && (m.CreatedAt.Before(cursor.CreatedAt) || (m.CreatedAt.Equal(cursor.CreatedAt) && id < cursor.MessageID)) {
				filtered = append(filtered, m)
			}
			if direction == "newer" && (m.CreatedAt.After(cursor.CreatedAt) || (m.CreatedAt.Equal(cursor.CreatedAt) && id > cursor.MessageID)) {
				filtered = append(filtered, m)
			}
		}
	}
	if limit > len(filtered) {
		limit = len(filtered)
	}
	return filtered[:limit], total, nil
}

func (r *MemoryRepository) SearchChats(ctx context.Context, params domain.SearchChatParams) (domain.SearchChatsResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	q := strings.ToLower(strings.TrimSpace(params.Query))
	if q == "" {
		return domain.SearchChatsResult{}, nil
	}

	hits := make([]domain.SearchChatHit, 0)
	for _, s := range r.sessions {
		if s.UserID != params.UserID || s.DeletedAt != nil {
			continue
		}
		titleMatched := strings.Contains(strings.ToLower(s.Title), q)

		var (
			matched      *domain.ChatMessage
			lastMessage  time.Time
		)
		for i := range r.messages[s.ID] {
			m := r.messages[s.ID][i]
			if m.CreatedAt.After(lastMessage) {
				lastMessage = m.CreatedAt
			}
			if m.Role != domain.RoleUser && m.Role != domain.RoleAssistant {
				continue
			}
			if strings.Contains(strings.ToLower(m.Content), q) {
				if matched == nil || m.CreatedAt.After(matched.CreatedAt) {
					cp := m
					matched = &cp
				}
			}
		}
		if !titleMatched && matched == nil {
			continue
		}
		updatedAt := lastMessage
		if updatedAt.IsZero() {
			updatedAt = s.CreatedAt
		}
		sortAt := updatedAt
		hit := domain.SearchChatHit{
			SessionID:        s.ID.String(),
			Title:            s.Title,
			SessionCreatedAt: s.CreatedAt,
			SessionUpdatedAt: updatedAt,
			LastMessageAt:    lastMessage,
			TitleMatched:     titleMatched,
			SortAt:           sortAt,
		}
		if matched != nil {
			hit.MatchedMessageID = matched.ID.String()
			hit.MatchedRole = matched.Role
			hit.MatchedContent = matched.Content
			hit.MatchedAt = matched.CreatedAt
			hit.SortAt = matched.CreatedAt
		}
		hits = append(hits, hit)
	}

	slices.SortFunc(hits, func(a, b domain.SearchChatHit) int {
		if a.SortAt.After(b.SortAt) {
			return -1
		}
		if a.SortAt.Before(b.SortAt) {
			return 1
		}
		if a.SessionID > b.SessionID {
			return -1
		}
		if a.SessionID < b.SessionID {
			return 1
		}
		return 0
	})

	total := len(hits)
	filtered := hits
	if params.Cursor != nil {
		filtered = make([]domain.SearchChatHit, 0, len(hits))
		for _, h := range hits {
			if h.SortAt.Before(params.Cursor.SortAt) || (h.SortAt.Equal(params.Cursor.SortAt) && h.SessionID < params.Cursor.SessionID) {
				filtered = append(filtered, h)
			}
		}
	}

	limit := params.Limit
	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}

	return domain.SearchChatsResult{
		TotalCount: total,
		Items:      filtered[:limit],
	}, nil
}
