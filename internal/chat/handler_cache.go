package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/messivite/go-thy-case-study-backend/internal/cache"
)

// HandlerOption configures Handler construction.
type HandlerOption func(*Handler)

// WithResponseCache enables GET /api/chats and GET /api/chats/{id}/messages response caching.
func WithResponseCache(store cache.Store, ttlList, ttlMsgs time.Duration) HandlerOption {
	return func(h *Handler) {
		if store != nil {
			h.respCache = store
		}
		h.ttlChatList = ttlList
		h.ttlChatMsgs = ttlMsgs
	}
}

func (h *Handler) cacheActive() bool {
	return h.ttlChatList > 0 || h.ttlChatMsgs > 0
}

func (h *Handler) tryWriteCachedJSON(w http.ResponseWriter, r *http.Request, key string, ttl time.Duration) bool {
	if !h.cacheActive() || ttl <= 0 {
		return false
	}
	b, ok := h.respCache.Get(r.Context(), key)
	if !ok || len(b) == 0 {
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
	return true
}

func (h *Handler) cachePutJSON(ctx context.Context, key string, ttl time.Duration, v any) {
	if !h.cacheActive() || ttl <= 0 {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.respCache.Set(ctx, key, b, ttl)
}

func (h *Handler) invalidateChatList(ctx context.Context, userID string) {
	if !h.cacheActive() {
		return
	}
	h.respCache.DeletePrefix(ctx, cache.PrefixChatList(userID))
}

func (h *Handler) invalidateChatMessages(ctx context.Context, userID, chatID string) {
	if !h.cacheActive() {
		return
	}
	h.respCache.DeletePrefix(ctx, cache.PrefixChatMessages(userID, chatID))
}

func (h *Handler) invalidateChatListAndSession(ctx context.Context, userID, chatID string) {
	h.invalidateChatList(ctx, userID)
	h.invalidateChatMessages(ctx, userID, chatID)
}
