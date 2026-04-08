package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	usecase "github.com/example/thy-case-study-backend/internal/application/chat"
	"github.com/example/thy-case-study-backend/internal/auth"
	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
	"github.com/example/thy-case-study-backend/internal/httpx"
)

type Handler struct {
	uc *usecase.UseCase
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

type createSessionRequest struct {
	Title string `json:"title"`
}

type createSessionResponse struct {
	ID string `json:"id"`
}

type postMessageRequest struct {
	Provider string        `json:"provider,omitempty"`
	Model    string        `json:"model,omitempty"`
	Content  string        `json:"content,omitempty"`
	Messages []chatMessage `json:"messages,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantResponse struct {
	AssistantMessage chatMessage    `json:"assistantMessage"`
	Usage            map[string]any `json:"usage,omitempty"`
}

type chatDetailResponse struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Messages []chatMessage `json:"messages"`
}

type chatListItemResponse struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	LastMessagePreview string `json:"lastMessagePreview"`
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

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	sessions, err := h.uc.ListSessions(r.Context(), user.UserID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	items := make([]chatListItemResponse, 0, len(sessions))
	for _, s := range sessions {
		lastPreview, updatedAt := h.uc.GetSessionSummary(r.Context(), s.ID.String())
		items = append(items, chatListItemResponse{
			ID:                 s.ID.String(),
			Title:              s.Title,
			CreatedAt:          s.CreatedAt.Format(timeFormat),
			UpdatedAt:          updatedAt.Format(timeFormat),
			LastMessagePreview: lastPreview,
		})
	}
	writeJSON(w, http.StatusOK, items)
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

	session, err := h.uc.CreateSession(r.Context(), user.UserID, req.Title)
	if err != nil {
		writeAppError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{ID: session.ID.String()})
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	messages, err := h.uc.GetSessionMessages(r.Context(), user.UserID, sessionID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIMessages(messages))
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

	resp := chatDetailResponse{
		ID:       session.ID.String(),
		Title:    session.Title,
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

	writeJSON(w, http.StatusCreated, assistantResponse{
		AssistantMessage: chatMessage{Role: string(assistantMsg.Role), Content: assistantMsg.Content},
		Usage:            usage,
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

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	streamCtx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	events, usage, finalize, err := h.uc.StreamMessage(streamCtx, user.UserID, chatID, req.Provider, req.Model, req.Content, toDomainMessages(req.Messages))
	if err != nil {
		writeAppError(w, err)
		return
	}

	meta := map[string]any{
		"provider": req.Provider,
		"model":    req.Model,
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
			return
		case ev, ok := <-events:
			if !ok {
				msg, ferr := finalize(out.String())
				if ferr != nil {
					_ = writeSSE(w, map[string]any{"type": "error", "message": ferr.Error()})
				} else {
					_ = writeSSE(w, map[string]any{"type": "meta", "meta": map[string]any{"assistantMessageId": msg.ID.String()}})
					_ = writeSSE(w, map[string]any{"type": "done"})
				}
				flusher.Flush()
				return
			}
			if ev.Type == domain.EventDelta {
				out.WriteString(ev.Delta)
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
		out = append(out, chatMessage{Role: string(m.Role), Content: m.Content})
	}
	return out
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
	case errors.Is(err, domain.ErrSessionNotFound):
		httpx.NotFound(w, err.Error())
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
	default:
		httpx.Internal(w)
	}
}
