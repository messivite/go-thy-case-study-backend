package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/cache"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/httpx"
)

type Handler struct {
	uc          *usecase.UseCase
	respCache   cache.Store
	ttlChatList time.Duration
	ttlChatMsgs time.Duration
}

func NewHandler(uc *usecase.UseCase, opts ...HandlerOption) *Handler {
	h := &Handler{uc: uc, respCache: cache.Nop{}}
	for _, o := range opts {
		o(h)
	}
	return h
}

type createSessionRequest struct {
	Title    string `json:"title"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	// Content doluysa oturum oluşturulduktan hemen sonra bu metinle LLM çağrılır;
	// yanıtta assistantMessage dönür (ayrıca POST /messages atmana gerek kalmaz).
	Content string `json:"content,omitempty"`
}

type createSessionResponse struct {
	ID               string         `json:"id"`
	Provider         string         `json:"provider"`
	Model            string         `json:"model"`
	AssistantMessage *chatMessage   `json:"assistantMessage,omitempty"`
	Usage            map[string]any `json:"usage,omitempty"`
}

type postMessageRequest struct {
	Provider string        `json:"provider,omitempty"`
	Model    string        `json:"model,omitempty"`
	Content  string        `json:"content,omitempty"`
	Messages []chatMessage `json:"messages,omitempty"`
}

type chatMessage struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

type syncRequest struct {
	Provider string        `json:"provider,omitempty"`
	Model    string        `json:"model,omitempty"`
	Messages []syncMessage `json:"messages"`
}

type syncMessage struct {
	Content string `json:"content"`
	SentAt  string `json:"sentAt,omitempty"`
}

type assistantResponse struct {
	AssistantMessage chatMessage    `json:"assistantMessage"`
	Usage            map[string]any `json:"usage,omitempty"`
}

type syncResponse struct {
	SyncedCount      int            `json:"syncedCount"`
	SyncedMessages   []chatMessage  `json:"syncedMessages"`
	AssistantMessage chatMessage    `json:"assistantMessage"`
	Usage            map[string]any `json:"usage,omitempty"`
}

type chatDetailResponse struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Provider string        `json:"provider"`
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatListItemResponse struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Provider           string `json:"provider"`
	Model              string `json:"model"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	LastMessagePreview string `json:"lastMessagePreview"`
}

type chatListPageResponse struct {
	TotalCount int                    `json:"totalCount"`
	HasNext    bool                   `json:"hasNext"`
	NextCursor string                 `json:"nextCursor,omitempty"`
	Items      []chatListItemResponse `json:"items"`
}

type messagesPageResponse struct {
	TotalCount int           `json:"totalCount"`
	HasNext    bool          `json:"hasNext"`
	NextCursor string        `json:"nextCursor,omitempty"`
	Direction  string        `json:"direction"`
	Items      []chatMessage `json:"items"`
}

type listProvidersResponse struct {
	Default   string             `json:"default"`
	Providers []providerInfoItem `json:"providers"`
}

type providerInfoItem struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

type supportedModelItem struct {
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	DisplayName    string `json:"displayName"`
	SupportsStream bool   `json:"supportsStream"`
}

type listModelsResponse struct {
	Models []supportedModelItem `json:"models"`
}

type searchChatsResponse struct {
	TotalCount int              `json:"totalCount"`
	HasNext    bool             `json:"hasNext"`
	NextCursor string           `json:"nextCursor,omitempty"`
	Items      []searchChatItem `json:"items"`
}

type searchChatItem struct {
	SessionID        string           `json:"sessionId"`
	Title            string           `json:"title"`
	SessionCreatedAt string           `json:"sessionCreatedAt"`
	SessionUpdatedAt string           `json:"sessionUpdatedAt"`
	LastMessageAt    string           `json:"lastMessageAt,omitempty"`
	TitleMatched     bool             `json:"titleMatched"`
	MatchedMessageID string           `json:"matchedMessageId,omitempty"`
	MatchedRole      string           `json:"matchedRole,omitempty"`
	MatchedContent   string           `json:"matchedContent,omitempty"`
	MatchedAt        string           `json:"matchedAt,omitempty"`
	Highlights       []highlightRange `json:"highlights,omitempty"`
}

type highlightRange struct {
	Field string `json:"field"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

const maxSyncMessages = 50

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	metas := h.uc.ListProviders()
	items := make([]providerInfoItem, 0, len(metas))
	for _, m := range metas {
		items = append(items, providerInfoItem{
			Name:    m.Name,
			Model:   m.DefaultModel,
			Enabled: m.Enabled,
		})
	}
	writeJSON(w, http.StatusOK, listProvidersResponse{
		Default:   h.uc.DefaultProvider(),
		Providers: items,
	})
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.AuthenticatedUserFromContext(r.Context()); !ok {
		httpx.Unauthorized(w)
		return
	}
	rows, err := h.uc.ListSupportedModels(r.Context())
	if err != nil {
		httpx.Internal(w)
		return
	}
	items := make([]supportedModelItem, 0, len(rows))
	for _, row := range rows {
		dn := row.DisplayName
		if strings.TrimSpace(dn) == "" {
			dn = row.ModelID
		}
		items = append(items, supportedModelItem{
			Provider:       row.Provider,
			Model:          row.ModelID,
			DisplayName:    dn,
			SupportsStream: row.SupportsStream,
		})
	}
	writeJSON(w, http.StatusOK, listModelsResponse{Models: items})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	prof, err := h.uc.GetUserProfile(r.Context(), user.UserID)
	if err != nil {
		httpx.Internal(w)
		return
	}
	writeJSON(w, http.StatusOK, toMeResponse(user, prof))
}

// meResponse is GET /api/me — JWT özeti (user) + public.profiles satırı (profile).
type meResponse struct {
	User    meUserSection    `json:"user"`
	Profile meProfileSection `json:"profile"`
}

type meUserSection struct {
	ID               string         `json:"id"`
	Email            string         `json:"email,omitempty"`
	Role             string         `json:"role,omitempty"`
	Roles            []string       `json:"roles,omitempty"`
	Phone            string         `json:"phone,omitempty"`
	SessionID        string         `json:"sessionId,omitempty"`
	Issuer           string         `json:"iss,omitempty"`
	Audience         string         `json:"aud,omitempty"`
	IssuedAt         int64          `json:"iat,omitempty"`
	ExpiresAt        int64          `json:"exp,omitempty"`
	IssuedAtRFC3339  string         `json:"issuedAt,omitempty"`
	ExpiresAtRFC3339 string         `json:"expiresAt,omitempty"`
	AppMetadata      map[string]any `json:"appMetadata,omitempty"`
	UserMetadata     map[string]any `json:"userMetadata,omitempty"`
}

type meProfileSection struct {
	ID                  string         `json:"id"`
	DisplayName         string         `json:"displayName,omitempty"`
	AvatarURL           string         `json:"avatarUrl,omitempty"`
	Role                string         `json:"role,omitempty"`
	IsActive            bool           `json:"isActive"`
	PreferredProvider   string         `json:"preferredProvider,omitempty"`
	PreferredModel      string         `json:"preferredModel,omitempty"`
	Locale              string         `json:"locale,omitempty"`
	Timezone            string         `json:"timezone,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	LastSeenAt          string         `json:"lastSeenAt,omitempty"`
	OnboardingCompleted bool           `json:"onboardingCompleted"`
	CreatedAt           string         `json:"createdAt,omitempty"`
	UpdatedAt           string         `json:"updatedAt,omitempty"`
	IsAnonymous         bool           `json:"isAnonymous"`
}

func toMeResponse(u *auth.AuthenticatedUser, p domain.UserProfile) meResponse {
	out := meResponse{
		User:    meUserSection{},
		Profile: toMeProfileSection(p),
	}
	if u != nil {
		out.User = meUserSection{
			ID:           u.UserID,
			Email:        u.Email,
			Role:         u.Role,
			Roles:        u.Roles,
			Phone:        u.Phone,
			SessionID:    u.SessionID,
			Issuer:       u.Issuer,
			Audience:     u.Audience,
			IssuedAt:     u.IssuedAt,
			ExpiresAt:    u.ExpiresAt,
			AppMetadata:  u.AppMetadata,
			UserMetadata: u.UserMetadata,
		}
		if u.IssuedAt > 0 {
			out.User.IssuedAtRFC3339 = time.Unix(u.IssuedAt, 0).UTC().Format(time.RFC3339)
		}
		if u.ExpiresAt > 0 {
			out.User.ExpiresAtRFC3339 = time.Unix(u.ExpiresAt, 0).UTC().Format(time.RFC3339)
		}
	}
	return out
}

func toMeProfileSection(p domain.UserProfile) meProfileSection {
	sec := meProfileSection{
		ID:                  p.ID,
		DisplayName:         p.DisplayName,
		AvatarURL:           p.AvatarURL,
		Role:                p.Role,
		IsActive:            p.IsActive,
		PreferredProvider:   p.PreferredProvider,
		PreferredModel:      p.PreferredModel,
		Locale:              p.Locale,
		Timezone:            p.Timezone,
		Metadata:            p.Metadata,
		OnboardingCompleted: p.OnboardingCompleted,
		IsAnonymous:         p.IsAnonymous,
	}
	if p.LastSeenAt != nil {
		sec.LastSeenAt = p.LastSeenAt.UTC().Format(time.RFC3339Nano)
	}
	if !p.CreatedAt.IsZero() {
		sec.CreatedAt = p.CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	if !p.UpdatedAt.IsZero() {
		sec.UpdatedAt = p.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return sec
}

type meUsageResponse struct {
	QuotaBypass bool            `json:"quotaBypass"`
	Daily       usageBucketJSON `json:"daily"`
	Weekly      usageBucketJSON `json:"weekly"`
	PeriodNote  string          `json:"periodNote"`
}

type usageBucketJSON struct {
	LimitTokens     int  `json:"limitTokens"`
	UsedTokens      int  `json:"usedTokens"`
	RemainingTokens *int `json:"remainingTokens,omitempty"`
}

func meUsagePeriodNote() string {
	return "Daily = UTC calendar day; weekly = rolling last 7 days (UTC). Counts only successful LLM completions (total_tokens)."
}

func toMeUsageResponse(m domain.MeUsage) meUsageResponse {
	return meUsageResponse{
		QuotaBypass: m.QuotaBypass,
		Daily:       toUsageBucketJSON(m.Daily),
		Weekly:      toUsageBucketJSON(m.Weekly),
		PeriodNote:  meUsagePeriodNote(),
	}
}

func toUsageBucketJSON(w domain.MeUsageWindow) usageBucketJSON {
	out := usageBucketJSON{
		LimitTokens: w.LimitTokens,
		UsedTokens:  w.UsedTokens,
	}
	if w.LimitTokens > 0 {
		rem := w.LimitTokens - w.UsedTokens
		if rem < 0 {
			rem = 0
		}
		out.RemainingTokens = &rem
	}
	return out
}

func (h *Handler) MeUsage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	mu, err := h.uc.MeUsage(r.Context(), user.UserID)
	if err != nil {
		httpx.Internal(w)
		return
	}
	writeJSON(w, http.StatusOK, toMeUsageResponse(mu))
}

func (h *Handler) SearchChats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	query := r.URL.Query().Get("q")
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			httpx.BadRequest(w, "invalid limit")
			return
		}
		limit = v
	}
	cursor := r.URL.Query().Get("cursor")

	result, err := h.uc.SearchChats(r.Context(), user.UserID, query, limit, cursor)
	if err != nil {
		writeAppError(w, err)
		return
	}
	items := make([]searchChatItem, 0, len(result.Items))
	for _, it := range result.Items {
		item := searchChatItem{
			SessionID:        it.SessionID,
			Title:            it.Title,
			SessionCreatedAt: it.SessionCreatedAt.Format(timeFormat),
			SessionUpdatedAt: it.SessionUpdatedAt.Format(timeFormat),
			TitleMatched:     it.TitleMatched,
		}
		if !it.LastMessageAt.IsZero() {
			item.LastMessageAt = it.LastMessageAt.Format(timeFormat)
		}
		if it.MatchedMessageID != "" {
			item.MatchedMessageID = it.MatchedMessageID
		}
		if it.MatchedRole != "" {
			item.MatchedRole = string(it.MatchedRole)
		}
		if it.MatchedContent != "" {
			item.MatchedContent = it.MatchedContent
		}
		if !it.MatchedAt.IsZero() {
			item.MatchedAt = it.MatchedAt.Format(timeFormat)
		}
		item.Highlights = collectHighlights(query, it.Title, it.MatchedContent)
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, searchChatsResponse{
		TotalCount: result.TotalCount,
		HasNext:    result.HasNext,
		NextCursor: result.NextCursor,
		Items:      items,
	})
}

func collectHighlights(query, title, content string) []highlightRange {
	out := make([]highlightRange, 0, 2)
	if s, e, ok := highlightBounds(title, query); ok {
		out = append(out, highlightRange{Field: "title", Start: s, End: e})
	}
	if s, e, ok := highlightBounds(content, query); ok {
		out = append(out, highlightRange{Field: "matchedContent", Start: s, End: e})
	}
	return out
}

func highlightBounds(text, query string) (int, int, bool) {
	t := strings.ToLower(strings.TrimSpace(text))
	q := strings.ToLower(strings.TrimSpace(query))
	if t == "" || q == "" {
		return 0, 0, false
	}
	idx := strings.Index(t, q)
	if idx < 0 {
		return 0, 0, false
	}
	return idx, idx + len(q), true
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	listKey := cache.KeyChatList(user.UserID, r.URL.RawQuery)
	if h.tryWriteCachedJSON(w, r, listKey, h.ttlChatList) {
		return
	}

	limit := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			httpx.BadRequest(w, "invalid limit")
			return
		}
		limit = v
	}
	cursor := r.URL.Query().Get("cursor")

	// Backward compatible: no query => legacy array response
	if limit == 0 && strings.TrimSpace(cursor) == "" {
		sessions, err := h.uc.ListSessions(r.Context(), user.UserID)
		if err != nil {
			writeAppError(w, err)
			return
		}
		items := make([]chatListItemResponse, 0, len(sessions))
		for _, s := range sessions {
			lastPreview, updatedAt := h.uc.GetSessionSummary(r.Context(), s.ID.String(), s.CreatedAt)
			lp, lm := sessionLLMSummary(s)
			items = append(items, chatListItemResponse{
				ID:                 s.ID.String(),
				Title:              s.Title,
				Provider:           lp,
				Model:              lm,
				CreatedAt:          s.CreatedAt.Format(timeFormat),
				UpdatedAt:          updatedAt.Format(timeFormat),
				LastMessagePreview: lastPreview,
			})
		}
		h.cachePutJSON(r.Context(), listKey, h.ttlChatList, items)
		writeJSON(w, http.StatusOK, items)
		return
	}

	page, err := h.uc.ListSessionsPage(r.Context(), user.UserID, limit, cursor)
	if err != nil {
		writeAppError(w, err)
		return
	}
	items := make([]chatListItemResponse, 0, len(page.Items))
	for _, it := range page.Items {
		s := it.Session
		lp, lm := sessionLLMSummary(s)
		items = append(items, chatListItemResponse{
			ID:                 s.ID.String(),
			Title:              s.Title,
			Provider:           lp,
			Model:              lm,
			CreatedAt:          s.CreatedAt.Format(timeFormat),
			UpdatedAt:          it.UpdatedAt.Format(timeFormat),
			LastMessagePreview: it.LastMessagePreview,
		})
	}
	body := chatListPageResponse{
		TotalCount: page.TotalCount,
		HasNext:    page.HasNext,
		NextCursor: page.NextCursor,
		Items:      items,
	}
	h.cachePutJSON(r.Context(), listKey, h.ttlChatList, body)
	writeJSON(w, http.StatusOK, body)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request payload")
		return
	}

	session, err := h.uc.CreateSession(r.Context(), user.UserID, req.Title, req.Provider, req.Model)
	if err != nil {
		writeAppError(w, err)
		return
	}

	chatID := session.ID.String()
	if strings.TrimSpace(req.Content) != "" {
		assistant, usage, sendErr := h.uc.SendMessage(r.Context(), user.UserID, chatID, req.Provider, req.Model, req.Content, nil)
		if sendErr != nil {
			writeAppError(w, sendErr)
			return
		}
		h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
		am := chatMessage{
			Role:     string(assistant.Role),
			Content:  assistant.Content,
			Provider: assistant.Provider,
			Model:    assistant.Model,
		}
		writeJSON(w, http.StatusCreated, createSessionResponse{
			ID:               chatID,
			Provider:         session.DefaultProvider,
			Model:            session.DefaultModel,
			AssistantMessage: &am,
			Usage:            usage,
		})
		return
	}

	h.invalidateChatList(r.Context(), user.UserID)
	writeJSON(w, http.StatusCreated, createSessionResponse{
		ID:       chatID,
		Provider: session.DefaultProvider,
		Model:    session.DefaultModel,
	})
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	chatID := chi.URLParam(r, "chatID")
	if err := h.uc.DeleteSession(r.Context(), user.UserID, chatID); err != nil {
		writeAppError(w, err)
		return
	}
	h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	chatID := chi.URLParam(r, "chatID")
	messageID := chi.URLParam(r, "messageID")
	if err := h.uc.DeleteOwnMessage(r.Context(), user.UserID, chatID, messageID); err != nil {
		writeAppError(w, err)
		return
	}
	h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	chatID := chi.URLParam(r, "chatID")
	msgKey := cache.KeyChatMessages(user.UserID, chatID, r.URL.RawQuery)
	if h.tryWriteCachedJSON(w, r, msgKey, h.ttlChatMsgs) {
		return
	}
	limit := 30
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			httpx.BadRequest(w, "invalid limit")
			return
		}
		limit = v
	}
	direction := strings.TrimSpace(r.URL.Query().Get("direction"))
	cursor := r.URL.Query().Get("cursor")

	page, err := h.uc.GetChatMessagesPage(r.Context(), user.UserID, chatID, limit, direction, cursor)
	if err != nil {
		writeAppError(w, err)
		return
	}
	body := messagesPageResponse{
		TotalCount: page.TotalCount,
		HasNext:    page.HasNext,
		NextCursor: page.NextCursor,
		Direction:  page.Direction,
		Items:      toAPIMessages(page.Messages),
	}
	h.cachePutJSON(r.Context(), msgKey, h.ttlChatMsgs, body)
	writeJSON(w, http.StatusOK, body)
}

func (h *Handler) GetChat(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	chatID := chi.URLParam(r, "chatID")
	session, messages, err := h.uc.GetChat(r.Context(), user.UserID, chatID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	rootP, rootM := lastAssistantLLMFromMessages(messages)
	if rootP == "" && rootM == "" {
		rootP, rootM = session.LastProvider, session.LastModel
	}
	if rootP == "" && rootM == "" {
		rootP, rootM = session.DefaultProvider, session.DefaultModel
	}
	resp := chatDetailResponse{
		ID:       session.ID.String(),
		Title:    session.Title,
		Provider: rootP,
		Model:    rootM,
		Messages: toAPIMessages(messages),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	chatID := chi.URLParam(r, "chatID")
	var req postMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request payload")
		return
	}

	assistantMsg, usage, err := h.uc.SendMessage(r.Context(), user.UserID, chatID, req.Provider, req.Model, req.Content, toDomainMessages(req.Messages))
	if err != nil {
		writeAppError(w, err)
		return
	}

	h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
	writeJSON(w, http.StatusCreated, assistantResponse{
		AssistantMessage: chatMessage{
			Role:     string(assistantMsg.Role),
			Content:  assistantMsg.Content,
			Provider: assistantMsg.Provider,
			Model:    assistantMsg.Model,
		},
		Usage: usage,
	})
}

func (h *Handler) StreamMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.Internal(w)
		return
	}

	chatID := chi.URLParam(r, "chatID")
	var req postMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request payload")
		return
	}

	streamCtx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	events, usage, finalize, cancelStream, err := h.uc.StreamMessage(streamCtx, user.UserID, chatID, req.Provider, req.Model, req.Content, toDomainMessages(req.Messages))
	if err != nil {
		writeAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	meta := map[string]any{
		"provider": usage["provider"],
		"model":    usage["model"],
		"usage":    usage,
	}
	_ = writeSSE(w, map[string]any{"type": "meta", "meta": meta})
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	var out strings.Builder
	for {
		select {
		case <-streamCtx.Done():
			partialRaw := out.String()
			if strings.TrimSpace(partialRaw) == "" {
				cancelStream(0)
			} else {
				msg, ferr := finalize(partialRaw)
				if ferr != nil {
					_ = writeSSE(w, map[string]any{"type": "error", "message": ferr.Error()})
				} else {
					_ = writeSSE(w, map[string]any{"type": "meta", "meta": map[string]any{
						"assistantMessageId": msg.ID.String(),
						"provider":           msg.Provider,
						"model":              msg.Model,
						"partial":            true,
					}})
					_ = writeSSE(w, map[string]any{"type": "cancelled"})
				}
				cancelStream(len(partialRaw))
			}
			h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
			flusher.Flush()
			return
		case ev, ok := <-events:
			if !ok {
				msg, ferr := finalize(out.String())
				if ferr != nil {
					_ = writeSSE(w, map[string]any{"type": "error", "message": ferr.Error()})
				} else {
					_ = writeSSE(w, map[string]any{"type": "meta", "meta": map[string]any{
						"assistantMessageId": msg.ID.String(),
						"provider":           msg.Provider,
						"model":              msg.Model,
					}})
					_ = writeSSE(w, map[string]any{"type": "done"})
				}
				h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
				flusher.Flush()
				return
			}
			if ev.Type == domain.EventDelta {
				out.WriteString(ev.Delta)
			}
			// Provider iç sinyali; kanal kapanınca zaten assistantMessageId + type "done" gönderilir.
			if ev.Type == domain.EventDone {
				continue
			}
			_ = writeSSE(w, map[string]any{
				"type":    string(ev.Type),
				"delta":   ev.Delta,
				"message": ev.Message,
				"meta":    ev.Meta,
			})
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handler) SyncMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	chatID := chi.URLParam(r, "chatID")
	var req syncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request payload")
		return
	}
	if len(req.Messages) == 0 {
		httpx.BadRequest(w, "messages cannot be empty")
		return
	}
	if len(req.Messages) > maxSyncMessages {
		httpx.BadRequest(w, fmt.Sprintf("messages exceeds limit (%d)", maxSyncMessages))
		return
	}

	type orderedMsg struct {
		content string
		sentAt  time.Time
		hasTime bool
		idx     int
	}
	ordered := make([]orderedMsg, 0, len(req.Messages))
	for i, msg := range req.Messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			httpx.BadRequest(w, "message content cannot be empty")
			return
		}
		item := orderedMsg{content: content, idx: i}
		if strings.TrimSpace(msg.SentAt) != "" {
			ts, err := time.Parse(time.RFC3339, msg.SentAt)
			if err != nil {
				httpx.BadRequest(w, "invalid sentAt format, expected RFC3339")
				return
			}
			item.sentAt = ts
			item.hasTime = true
		}
		ordered = append(ordered, item)
	}
	slices.SortStableFunc(ordered, func(a, b orderedMsg) int {
		if a.hasTime && b.hasTime {
			if a.sentAt.Before(b.sentAt) {
				return -1
			}
			if a.sentAt.After(b.sentAt) {
				return 1
			}
		}
		return a.idx - b.idx
	})

	batch := make([]domain.BatchMessage, 0, len(ordered))
	for _, item := range ordered {
		batch = append(batch, domain.BatchMessage{
			Content: item.content,
		})
	}

	result, err := h.uc.SyncMessages(r.Context(), user.UserID, chatID, req.Provider, req.Model, batch)
	if err != nil {
		writeAppError(w, err)
		return
	}

	h.invalidateChatListAndSession(r.Context(), user.UserID, chatID)
	writeJSON(w, http.StatusCreated, syncResponse{
		SyncedCount:      len(result.SyncedMessages),
		SyncedMessages:   toAPIMessages(result.SyncedMessages),
		AssistantMessage: toAPIMessages([]domain.ChatMessage{result.AssistantMessage})[0],
		Usage:            result.Usage,
	})
}

func toDomainMessages(messages []chatMessage) []domain.ChatMessage {
	out := make([]domain.ChatMessage, 0, len(messages))
	for _, m := range messages {
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		out = append(out, domain.ChatMessage{
			Role:    domain.Role(m.Role),
			Content: m.Content,
		})
	}
	return out
}

func toAPIMessages(messages []domain.ChatMessage) []chatMessage {
	out := make([]chatMessage, 0, len(messages))
	for _, m := range messages {
		out = append(out, chatMessage{
			Role:     string(m.Role),
			Content:  m.Content,
			Provider: m.Provider,
			Model:    m.Model,
		})
	}
	return out
}

func lastAssistantLLMFromMessages(msgs []domain.ChatMessage) (provider, model string) {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != domain.RoleAssistant {
			continue
		}
		if msgs[i].Provider != "" || msgs[i].Model != "" {
			return msgs[i].Provider, msgs[i].Model
		}
	}
	return "", ""
}

// sessionLLMSummary: son LLM turu (last_*) yoksa oturum default’ları (default_*).
func sessionLLMSummary(s domain.ChatSession) (provider, model string) {
	p, m := s.LastProvider, s.LastModel
	if p == "" && m == "" {
		p, m = s.DefaultProvider, s.DefaultModel
	}
	return p, m
}

func writeSSE(w http.ResponseWriter, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", b)
	return err
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		httpx.Forbidden(w)
	case errors.Is(err, domain.ErrMissingContent), errors.Is(err, domain.ErrInvalidRole):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrInvalidSessionID), errors.Is(err, domain.ErrInvalidMessageID), errors.Is(err, domain.ErrInvalidSearchCursor), errors.Is(err, domain.ErrSearchQueryTooShort), errors.Is(err, domain.ErrInvalidDirection):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrSessionNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrMessageNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrModelDiscontinued):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrUnsupportedProvider):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrProviderAuthFailed):
		httpx.ProviderAuthFailed(w)
	case errors.Is(err, domain.ErrProviderTimeout):
		httpx.ProviderTimeout(w)
	case errors.Is(err, domain.ErrProviderRateLimited):
		httpx.ProviderRateLimited(w)
	case errors.Is(err, domain.ErrProviderUnavailable):
		httpx.ProviderUnavailable(w)
	case errors.Is(err, domain.ErrProviderBadRequest):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrQuotaDailyExceeded):
		httpx.QuotaDailyExceeded(w)
	case errors.Is(err, domain.ErrQuotaWeeklyExceeded):
		httpx.QuotaWeeklyExceeded(w)
	case errors.Is(err, domain.ErrUserCancelled):
		httpx.GenerationCancelled(w)
	default:
		httpx.Internal(w)
	}
}
