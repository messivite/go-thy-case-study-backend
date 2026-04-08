package chat

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/example/thy-case-study-backend/internal/auth"
)

type Handler struct {
	service *ChatService
}

func NewHandler(service *ChatService) *Handler {
	return &Handler{service: service}
}

type createSessionRequest struct {
	Title string `json:"title"`
}

type createSessionResponse struct {
	ID string `json:"id"`
}

type postMessageRequest struct {
	Provider string `json:"provider,omitempty"`
	Content  string `json:"content"`
}

type listProvidersResponse struct {
	Providers []string `json:"providers"`
}

func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.service.providerFactory.ListProviders()
	writeJSON(w, http.StatusOK, listProvidersResponse{Providers: providers})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessions, err := h.service.ListSessions(r.Context(), user.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	session, err := h.service.CreateSession(r.Context(), user.UserID, req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{ID: session.ID.String()})
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	messages, err := h.service.GetSessionMessages(r.Context(), user.UserID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := chi.URLParam(r, "sessionID")
	var req postMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	_, assistantMessage, err := h.service.SendMessage(r.Context(), user.UserID, sessionID, req.Provider, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, assistantMessage)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
