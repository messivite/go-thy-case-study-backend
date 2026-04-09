package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

var _ domain.Repository = (*SupabaseRepository)(nil)

type SupabaseRepository struct {
	baseURL        string
	serviceRoleKey string
	client         *http.Client
}

func NewSupabaseRepository(supabaseURL, serviceRoleKey string) *SupabaseRepository {
	return &SupabaseRepository{
		baseURL:        supabaseURL + "/rest/v1",
		serviceRoleKey: serviceRoleKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *SupabaseRepository) CreateChatSession(ctx context.Context, userID, title, defaultProvider, defaultModel string) (domain.ChatSession, error) {
	body := map[string]any{
		"user_id": userID,
		"title":   title,
	}
	if defaultProvider != "" {
		body["default_provider"] = defaultProvider
	}
	if defaultModel != "" {
		body["default_model"] = defaultModel
	}

	var rows []chatSessionRow
	if err := r.doRequest(ctx, http.MethodPost, "/chat_sessions", body, &rows); err != nil {
		return domain.ChatSession{}, fmt.Errorf("create session: %w", err)
	}

	if len(rows) == 0 {
		return domain.ChatSession{}, fmt.Errorf("create session: no rows returned")
	}

	return rows[0].toDomain(), nil
}

func (r *SupabaseRepository) GetChatSessionsByUser(ctx context.Context, userID string) ([]domain.ChatSession, error) {
	path := fmt.Sprintf("/chat_sessions?user_id=eq.%s&order=created_at.desc", userID)

	var rows []chatSessionRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	sessions := make([]domain.ChatSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, row.toDomain())
	}
	return sessions, nil
}

func (r *SupabaseRepository) GetChatSessionByID(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return domain.ChatSession{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	path := fmt.Sprintf("/chat_sessions?id=eq.%s", sessionID)

	var rows []chatSessionRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return domain.ChatSession{}, fmt.Errorf("get session: %w", err)
	}

	if len(rows) == 0 {
		return domain.ChatSession{}, domain.ErrSessionNotFound
	}

	return rows[0].toDomain(), nil
}

func (r *SupabaseRepository) UpdateSessionLastLLM(ctx context.Context, sessionID, provider, model string) error {
	if _, err := uuid.Parse(sessionID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	body := map[string]any{
		"last_provider": provider,
		"last_model":    model,
	}
	path := fmt.Sprintf("/chat_sessions?id=eq.%s", sessionID)
	return r.doRequest(ctx, http.MethodPatch, path, body, nil)
}

func (r *SupabaseRepository) SaveMessage(ctx context.Context, sessionID, userID string, role domain.Role, content, provider, model string) (domain.ChatMessage, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return domain.ChatMessage{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	body := map[string]any{
		"session_id": sessionID,
		"role":       string(role),
		"content":    content,
	}
	if userID != "" {
		body["user_id"] = userID
	}
	if provider != "" {
		body["provider"] = provider
	}
	if model != "" {
		body["model"] = model
	}

	var rows []chatMessageRow
	if err := r.doRequest(ctx, http.MethodPost, "/chat_messages", body, &rows); err != nil {
		return domain.ChatMessage{}, fmt.Errorf("save message: %w", err)
	}

	if len(rows) == 0 {
		return domain.ChatMessage{}, fmt.Errorf("save message: no rows returned")
	}

	return rows[0].toDomain(), nil
}

func (r *SupabaseRepository) SaveMessages(ctx context.Context, sessionID, userID string, messages []domain.BatchMessage) ([]domain.ChatMessage, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	body := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		row := map[string]any{
			"session_id": sessionID,
			"user_id":    userID,
			"role":       string(domain.RoleUser),
			"content":    msg.Content,
		}
		if msg.Provider != "" {
			row["provider"] = msg.Provider
		}
		if msg.Model != "" {
			row["model"] = msg.Model
		}
		body = append(body, row)
	}

	var rows []chatMessageRow
	if err := r.doRequest(ctx, http.MethodPost, "/chat_messages", body, &rows); err != nil {
		return nil, fmt.Errorf("save messages: %w", err)
	}
	if len(rows) == 0 {
		return []domain.ChatMessage{}, nil
	}

	saved := make([]domain.ChatMessage, 0, len(rows))
	for _, row := range rows {
		saved = append(saved, row.toDomain())
	}
	return saved, nil
}

func (r *SupabaseRepository) GetMessagesBySession(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	path := fmt.Sprintf("/chat_messages?session_id=eq.%s&order=created_at.asc", sessionID)

	var rows []chatMessageRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	messages := make([]domain.ChatMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, row.toDomain())
	}
	return messages, nil
}

// ---------------------------------------------------------------------------
// HTTP helper
// ---------------------------------------------------------------------------

func (r *SupabaseRepository) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", r.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+r.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("supabase request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w (body=%s)", err, string(respBody))
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Row types for JSON mapping
// ---------------------------------------------------------------------------

type chatSessionRow struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Title            string  `json:"title"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	LastProvider     *string `json:"last_provider"`
	LastModel        *string `json:"last_model"`
	DefaultProvider  *string `json:"default_provider"`
	DefaultModel     *string `json:"default_model"`
}

func (r chatSessionRow) toDomain() domain.ChatSession {
	id, _ := uuid.Parse(r.ID)
	createdAt, _ := time.Parse(time.RFC3339Nano, r.CreatedAt)
	s := domain.ChatSession{
		ID:        id,
		UserID:    r.UserID,
		Title:     r.Title,
		CreatedAt: createdAt,
	}
	if r.LastProvider != nil {
		s.LastProvider = *r.LastProvider
	}
	if r.LastModel != nil {
		s.LastModel = *r.LastModel
	}
	if r.DefaultProvider != nil {
		s.DefaultProvider = *r.DefaultProvider
	}
	if r.DefaultModel != nil {
		s.DefaultModel = *r.DefaultModel
	}
	return s
}

type chatMessageRow struct {
	ID        string  `json:"id"`
	SessionID string  `json:"session_id"`
	UserID    *string `json:"user_id"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
	Provider  *string `json:"provider"`
	Model     *string `json:"model"`
}

func (r chatMessageRow) toDomain() domain.ChatMessage {
	id, _ := uuid.Parse(r.ID)
	sessionID, _ := uuid.Parse(r.SessionID)
	createdAt, _ := time.Parse(time.RFC3339Nano, r.CreatedAt)

	userID := ""
	if r.UserID != nil {
		userID = *r.UserID
	}

	m := domain.ChatMessage{
		ID:        id,
		SessionID: sessionID,
		UserID:    userID,
		Role:      domain.Role(r.Role),
		Content:   r.Content,
		CreatedAt: createdAt,
	}
	if r.Provider != nil {
		m.Provider = *r.Provider
	}
	if r.Model != nil {
		m.Model = *r.Model
	}
	return m
}
