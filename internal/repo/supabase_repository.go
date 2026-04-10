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
	path := fmt.Sprintf("/chat_sessions?user_id=eq.%s&deleted_at=is.null&order=created_at.desc", userID)

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

func (r *SupabaseRepository) GetChatSessionsByUserPage(ctx context.Context, userID string, limit int, cursor *domain.SessionCursor) (domain.SessionListPage, error) {
	body := map[string]any{
		"p_user_id": userID,
		"p_limit":   limit,
	}
	if cursor != nil {
		body["p_cursor_sort_at"] = cursor.SortAt.UTC().Format(time.RFC3339Nano)
		body["p_cursor_session_id"] = cursor.SessionID
	}
	var rows []sessionPageRow
	if err := r.doRequest(ctx, http.MethodPost, "/rpc/llm_get_user_chat_sessions_page", body, &rows); err != nil {
		return domain.SessionListPage{}, fmt.Errorf("list session page: %w", err)
	}
	if len(rows) == 0 {
		return domain.SessionListPage{}, nil
	}
	items := make([]domain.SessionListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return domain.SessionListPage{
		TotalCount: rows[0].TotalCount,
		Items:      items,
	}, nil
}

func (r *SupabaseRepository) GetChatSessionByID(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return domain.ChatSession{}, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	path := fmt.Sprintf("/chat_sessions?id=eq.%s&deleted_at=is.null", sessionID)

	var rows []chatSessionRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return domain.ChatSession{}, fmt.Errorf("get session: %w", err)
	}

	if len(rows) == 0 {
		return domain.ChatSession{}, domain.ErrSessionNotFound
	}

	return rows[0].toDomain(), nil
}

func (r *SupabaseRepository) SoftDeleteChatSession(ctx context.Context, sessionID string) error {
	if _, err := uuid.Parse(sessionID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	body := map[string]any{"deleted_at": now}
	path := fmt.Sprintf("/chat_sessions?id=eq.%s&deleted_at=is.null", sessionID)
	var rows []chatSessionRow
	if err := r.doRequest(ctx, http.MethodPatch, path, body, &rows); err != nil {
		return fmt.Errorf("soft delete session: %w", err)
	}
	if len(rows) == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *SupabaseRepository) UpdateSessionLastLLM(ctx context.Context, sessionID, provider, model string) error {
	if _, err := uuid.Parse(sessionID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	body := map[string]any{
		"last_provider": provider,
		"last_model":    model,
	}
	path := fmt.Sprintf("/chat_sessions?id=eq.%s&deleted_at=is.null", sessionID)
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

func (r *SupabaseRepository) SoftDeleteUserMessage(ctx context.Context, sessionID, messageID, userID string) error {
	if _, err := uuid.Parse(sessionID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}
	if _, err := uuid.Parse(messageID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidMessageID, err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	body := map[string]any{"deleted_at": now}
	path := fmt.Sprintf("/chat_messages?id=eq.%s&session_id=eq.%s&user_id=eq.%s&role=eq.user&deleted_at=is.null", messageID, sessionID, userID)
	var rows []chatMessageRow
	if err := r.doRequest(ctx, http.MethodPatch, path, body, &rows); err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	if len(rows) == 0 {
		return domain.ErrMessageNotFound
	}
	return nil
}

func (r *SupabaseRepository) GetMessagesBySession(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}

	path := fmt.Sprintf("/chat_messages?session_id=eq.%s&deleted_at=is.null&order=created_at.asc", sessionID)

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

func (r *SupabaseRepository) GetMessagesBySessionPage(ctx context.Context, sessionID string, limit int, direction string, cursor *domain.MessageCursor) ([]domain.ChatMessage, int, error) {
	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrInvalidSessionID, err)
	}
	if direction == "" {
		direction = "older"
	}
	if direction != "older" && direction != "newer" {
		return nil, 0, domain.ErrInvalidDirection
	}
	body := map[string]any{
		"p_session_id": sessionID,
		"p_limit":      limit,
		"p_direction":  direction,
	}
	if cursor != nil {
		body["p_cursor_created_at"] = cursor.CreatedAt.UTC().Format(time.RFC3339Nano)
		body["p_cursor_message_id"] = cursor.MessageID
	}
	var rows []messagePageRow
	if err := r.doRequest(ctx, http.MethodPost, "/rpc/llm_get_session_messages_page", body, &rows); err != nil {
		return nil, 0, fmt.Errorf("messages page: %w", err)
	}
	if len(rows) == 0 {
		return []domain.ChatMessage{}, 0, nil
	}
	out := make([]domain.ChatMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomainMessage())
	}
	return out, rows[0].TotalCount, nil
}

func (r *SupabaseRepository) SearchChats(ctx context.Context, params domain.SearchChatParams) (domain.SearchChatsResult, error) {
	body := map[string]any{
		"p_user_id": params.UserID,
		"p_query":   params.Query,
		"p_limit":   params.Limit,
	}
	if params.Cursor != nil {
		body["p_cursor_sort_at"] = params.Cursor.SortAt.UTC().Format(time.RFC3339Nano)
		body["p_cursor_session_id"] = params.Cursor.SessionID
	}

	var rows []searchChatRow
	if err := r.doRequest(ctx, http.MethodPost, "/rpc/llm_search_user_chats", body, &rows); err != nil {
		return domain.SearchChatsResult{}, fmt.Errorf("search chats: %w", err)
	}
	if len(rows) == 0 {
		return domain.SearchChatsResult{}, nil
	}
	items := make([]domain.SearchChatHit, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return domain.SearchChatsResult{
		TotalCount: rows[0].TotalCount,
		Items:      items,
	}, nil
}

func (r *SupabaseRepository) GetUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return domain.UserProfile{}, fmt.Errorf("get profile: %w", err)
	}
	path := "/profiles?id=eq." + userID + "&select=*"
	var rows []profileRow
	if err := r.doRequest(ctx, http.MethodGet, path, nil, &rows); err != nil {
		return domain.UserProfile{}, fmt.Errorf("get profile: %w", err)
	}
	if len(rows) == 0 {
		return domain.UserProfile{ID: userID, Locale: "tr"}, nil
	}
	return rows[0].toDomain()
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

type profileRow struct {
	ID                  string          `json:"id"`
	DisplayName         *string         `json:"display_name"`
	AvatarURL           *string         `json:"avatar_url"`
	Role                string          `json:"role"`
	IsActive            bool            `json:"is_active"`
	PreferredProvider   *string         `json:"preferred_provider"`
	PreferredModel      *string         `json:"preferred_model"`
	Locale              string          `json:"locale"`
	Timezone            *string         `json:"timezone"`
	Metadata            json.RawMessage `json:"metadata"`
	LastSeenAt          *string         `json:"last_seen_at"`
	OnboardingCompleted bool            `json:"onboarding_completed"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
	IsAnonymous         bool            `json:"is_anonymous"`
}

func (r profileRow) toDomain() (domain.UserProfile, error) {
	out := domain.UserProfile{
		ID:                  r.ID,
		Role:                r.Role,
		IsActive:            r.IsActive,
		Locale:              r.Locale,
		OnboardingCompleted: r.OnboardingCompleted,
		IsAnonymous:         r.IsAnonymous,
	}
	if r.DisplayName != nil {
		out.DisplayName = *r.DisplayName
	}
	if r.AvatarURL != nil {
		out.AvatarURL = *r.AvatarURL
	}
	if r.PreferredProvider != nil {
		out.PreferredProvider = *r.PreferredProvider
	}
	if r.PreferredModel != nil {
		out.PreferredModel = *r.PreferredModel
	}
	if r.Timezone != nil {
		out.Timezone = *r.Timezone
	}
	var err error
	out.CreatedAt, err = time.Parse(time.RFC3339Nano, r.CreatedAt)
	if err != nil {
		out.CreatedAt = time.Time{}
	}
	out.UpdatedAt, err = time.Parse(time.RFC3339Nano, r.UpdatedAt)
	if err != nil {
		out.UpdatedAt = time.Time{}
	}
	if r.LastSeenAt != nil && *r.LastSeenAt != "" {
		t, e := time.Parse(time.RFC3339Nano, *r.LastSeenAt)
		if e == nil {
			out.LastSeenAt = &t
		}
	}
	if len(r.Metadata) > 0 && string(r.Metadata) != "null" {
		var m map[string]any
		if json.Unmarshal(r.Metadata, &m) == nil && m != nil {
			out.Metadata = m
		}
	}
	return out, nil
}

type chatSessionRow struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Title            string  `json:"title"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	DeletedAt        *string `json:"deleted_at"`
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
	if r.DeletedAt != nil {
		t, err := time.Parse(time.RFC3339Nano, *r.DeletedAt)
		if err == nil {
			s.DeletedAt = &t
		}
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
	DeletedAt *string `json:"deleted_at"`
	Provider  *string `json:"provider"`
	Model     *string `json:"model"`
}

type searchChatRow struct {
	TotalCount       int     `json:"total_count"`
	SessionID        string  `json:"session_id"`
	Title            string  `json:"title"`
	SessionCreatedAt string  `json:"session_created_at"`
	SessionUpdatedAt string  `json:"session_updated_at"`
	LastMessageAt    *string `json:"last_message_at"`
	TitleMatched     bool    `json:"title_matched"`
	MatchedMessageID *string `json:"matched_message_id"`
	MatchedRole      *string `json:"matched_role"`
	MatchedContent   *string `json:"matched_content"`
	MatchedAt        *string `json:"matched_at"`
	SortAt           string  `json:"sort_at"`
}

type sessionPageRow struct {
	TotalCount        int     `json:"total_count"`
	SessionID         string  `json:"session_id"`
	Title             string  `json:"title"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DefaultProvider   *string `json:"default_provider"`
	DefaultModel      *string `json:"default_model"`
	LastProvider      *string `json:"last_provider"`
	LastModel         *string `json:"last_model"`
	LastMessagePreview *string `json:"last_message_preview"`
	SortAt            string  `json:"sort_at"`
}

func (r sessionPageRow) toDomain() domain.SessionListItem {
	createdAt, _ := time.Parse(time.RFC3339Nano, r.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, r.UpdatedAt)
	sortAt, _ := time.Parse(time.RFC3339Nano, r.SortAt)
	sid, _ := uuid.Parse(r.SessionID)
	s := domain.ChatSession{
		ID:        sid,
		Title:     r.Title,
		CreatedAt: createdAt,
	}
	if r.DefaultProvider != nil {
		s.DefaultProvider = *r.DefaultProvider
	}
	if r.DefaultModel != nil {
		s.DefaultModel = *r.DefaultModel
	}
	if r.LastProvider != nil {
		s.LastProvider = *r.LastProvider
	}
	if r.LastModel != nil {
		s.LastModel = *r.LastModel
	}
	preview := ""
	if r.LastMessagePreview != nil {
		preview = *r.LastMessagePreview
	}
	return domain.SessionListItem{
		Session:            s,
		LastMessagePreview: preview,
		UpdatedAt:          updatedAt,
		SortAt:             sortAt,
	}
}

type messagePageRow struct {
	TotalCount int     `json:"total_count"`
	ID         string  `json:"id"`
	SessionID  string  `json:"session_id"`
	UserID     *string `json:"user_id"`
	Role       string  `json:"role"`
	Content    string  `json:"content"`
	CreatedAt  string  `json:"created_at"`
	Provider   *string `json:"provider"`
	Model      *string `json:"model"`
}

func (r messagePageRow) toDomainMessage() domain.ChatMessage {
	row := chatMessageRow{
		ID:        r.ID,
		SessionID: r.SessionID,
		UserID:    r.UserID,
		Role:      r.Role,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		Provider:  r.Provider,
		Model:     r.Model,
	}
	return row.toDomain()
}

func (r searchChatRow) toDomain() domain.SearchChatHit {
	sessionCreatedAt, _ := time.Parse(time.RFC3339Nano, r.SessionCreatedAt)
	sessionUpdatedAt, _ := time.Parse(time.RFC3339Nano, r.SessionUpdatedAt)
	sortAt, _ := time.Parse(time.RFC3339Nano, r.SortAt)
	out := domain.SearchChatHit{
		SessionID:        r.SessionID,
		Title:            r.Title,
		SessionCreatedAt: sessionCreatedAt,
		SessionUpdatedAt: sessionUpdatedAt,
		TitleMatched:     r.TitleMatched,
		SortAt:           sortAt,
	}
	if r.LastMessageAt != nil {
		out.LastMessageAt, _ = time.Parse(time.RFC3339Nano, *r.LastMessageAt)
	}
	if r.MatchedMessageID != nil {
		out.MatchedMessageID = *r.MatchedMessageID
	}
	if r.MatchedRole != nil {
		out.MatchedRole = domain.Role(*r.MatchedRole)
	}
	if r.MatchedContent != nil {
		out.MatchedContent = *r.MatchedContent
	}
	if r.MatchedAt != nil {
		out.MatchedAt, _ = time.Parse(time.RFC3339Nano, *r.MatchedAt)
	}
	return out
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
	if r.DeletedAt != nil {
		t, err := time.Parse(time.RFC3339Nano, *r.DeletedAt)
		if err == nil {
			m.DeletedAt = &t
		}
	}
	if r.Provider != nil {
		m.Provider = *r.Provider
	}
	if r.Model != nil {
		m.Model = *r.Model
	}
	return m
}
