package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/messivite/go-thy-case-study-backend/internal/catalog"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/observability"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

type StreamFinalize func(assistantContent string, streamUsage map[string]any) (domain.ChatMessage, error)
type StreamCancel func(partialChars int)

type SyncResult struct {
	SyncedMessages   []domain.ChatMessage
	AssistantMessage domain.ChatMessage
	Usage            map[string]any
}

type SearchResult struct {
	TotalCount int
	Items      []domain.SearchChatHit
	NextCursor string
	HasNext    bool
}

type ChatMessagesPage struct {
	TotalCount int
	Messages   []domain.ChatMessage
	NextCursor string
	HasNext    bool
	Direction  string
}

type UseCase struct {
	repo      domain.Repository
	quotaRepo domain.QuotaRepository
	registry  *provider.Registry
	models    domain.SupportedModelsCatalog
}

func NewUseCase(repo domain.Repository, quotaRepo domain.QuotaRepository, registry *provider.Registry, models domain.SupportedModelsCatalog) *UseCase {
	return &UseCase{repo: repo, quotaRepo: quotaRepo, registry: registry, models: models}
}

func (uc *UseCase) GetUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	return uc.repo.GetUserProfile(ctx, userID)
}

const maxAvatarUploadBytes = 25 << 20 // raw bytes before resize (guardrail)

func profilePatchIsEmpty(p domain.ProfilePatch) bool {
	return p.DisplayName == nil && p.PreferredProvider == nil && p.PreferredModel == nil &&
		p.Locale == nil && p.Timezone == nil && p.AvatarURL == nil && p.OnboardingCompleted == nil
}

// PatchMe applies partial profile updates; rawAvatar (if non-empty) is resized to 300×300 JPEG, uploaded, and avatar_url is set.
func (uc *UseCase) PatchMe(ctx context.Context, userID string, patch domain.ProfilePatch, rawAvatar []byte) (domain.UserProfile, error) {
	if len(rawAvatar) > maxAvatarUploadBytes {
		return domain.UserProfile{}, domain.ErrAvatarTooLarge
	}
	if len(rawAvatar) > 0 {
		jpeg, err := repo.ResizeToAvatarJPEG(rawAvatar)
		if err != nil {
			return domain.UserProfile{}, fmt.Errorf("%w: %v", domain.ErrInvalidImagePayload, err)
		}
		url, err := uc.repo.UploadUserAvatarJPEG(ctx, userID, jpeg)
		if err != nil {
			return domain.UserProfile{}, err
		}
		patch.AvatarURL = &url
	}
	if profilePatchIsEmpty(patch) {
		return domain.UserProfile{}, domain.ErrProfilePatchEmpty
	}
	return uc.repo.PatchUserProfile(ctx, userID, patch)
}

func (uc *UseCase) CreateSession(ctx context.Context, userID, title, provider, model string) (domain.ChatSession, error) {
	dp := strings.TrimSpace(provider)
	dm := strings.TrimSpace(model)
	if dp == "" {
		dp = uc.registry.Default()
	}
	if dm == "" {
		if meta, ok := uc.registry.Meta(dp); ok {
			dm = strings.TrimSpace(meta.DefaultModel)
		}
	}
	eff := effectiveLLMModel(uc.registry, dp, model, "")
	if err := uc.ensureModelActive(ctx, dp, eff); err != nil {
		return domain.ChatSession{}, err
	}
	return uc.repo.CreateChatSession(ctx, userID, title, dp, dm)
}

func (uc *UseCase) ListSessions(ctx context.Context, userID string) ([]domain.ChatSession, error) {
	return uc.repo.GetChatSessionsByUser(ctx, userID)
}

func (uc *UseCase) GetChat(ctx context.Context, userID, chatID string) (domain.ChatSession, []domain.ChatMessage, error) {
	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return domain.ChatSession{}, nil, err
	}
	if session.UserID != userID {
		return domain.ChatSession{}, nil, domain.ErrUnauthorized
	}
	messages, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		return domain.ChatSession{}, nil, err
	}
	if err := uc.attachMessageLiked(ctx, userID, messages); err != nil {
		return domain.ChatSession{}, nil, err
	}
	return session, messages, nil
}

func (uc *UseCase) GetSessionMessages(ctx context.Context, userID, sessionID string) ([]domain.ChatMessage, error) {
	session, err := uc.repo.GetChatSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, domain.ErrUnauthorized
	}
	return uc.repo.GetMessagesBySession(ctx, sessionID)
}

func (uc *UseCase) DeleteSession(ctx context.Context, userID, chatID string) error {
	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return domain.ErrUnauthorized
	}
	return uc.repo.SoftDeleteChatSession(ctx, chatID)
}

func (uc *UseCase) DeleteOwnMessage(ctx context.Context, userID, chatID, messageID string) error {
	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return domain.ErrUnauthorized
	}
	return uc.repo.SoftDeleteUserMessage(ctx, chatID, messageID, userID)
}

// SetChatMessageLike: action 1 = like, 2 = unlike. Returns state 1 = liked, 2 = unliked (tek Supabase RPC).
func (uc *UseCase) SetChatMessageLike(ctx context.Context, userID, chatID, messageID string, action int) (state int, err error) {
	return uc.repo.SetChatMessageLike(ctx, userID, chatID, messageID, action)
}

func ptrBool(b bool) *bool {
	x := b
	return &x
}

// attachMessageLiked fills Liked for API: system → nil; user/assistant → false/true (toplu repo sorgusu).
func (uc *UseCase) attachMessageLiked(ctx context.Context, userID string, msgs []domain.ChatMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(msgs))
	for i := range msgs {
		m := msgs[i]
		if m.DeletedAt != nil {
			continue
		}
		if m.Role == domain.RoleSystem {
			continue
		}
		if m.ID == uuid.Nil {
			continue
		}
		ids = append(ids, m.ID.String())
	}
	likedMap, err := uc.repo.MessageLikedByUser(ctx, userID, ids)
	if err != nil {
		return err
	}
	for i := range msgs {
		m := &msgs[i]
		if m.Role == domain.RoleSystem {
			m.Liked = nil
			continue
		}
		if m.ID == uuid.Nil || m.DeletedAt != nil {
			m.Liked = nil
			continue
		}
		m.Liked = ptrBool(likedMap[m.ID.String()])
	}
	return nil
}

func (uc *UseCase) GetSessionSummary(ctx context.Context, chatID string, sessionCreatedAt time.Time) (lastMessagePreview string, updatedAt time.Time) {
	msgs, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil || len(msgs) == 0 {
		return "", sessionCreatedAt
	}
	// Newest non-deleted message drives updatedAt; preview skips empty content (e.g. streaming placeholder before cancel).
	var lastActivity *domain.ChatMessage
	var lastNonEmpty *domain.ChatMessage
	for i := range msgs {
		m := &msgs[i]
		if m.DeletedAt != nil {
			continue
		}
		if lastActivity == nil ||
			m.CreatedAt.After(lastActivity.CreatedAt) ||
			(m.CreatedAt.Equal(lastActivity.CreatedAt) && m.ID.String() > lastActivity.ID.String()) {
			lastActivity = m
		}
		if strings.TrimSpace(m.Content) != "" {
			if lastNonEmpty == nil ||
				m.CreatedAt.After(lastNonEmpty.CreatedAt) ||
				(m.CreatedAt.Equal(lastNonEmpty.CreatedAt) && m.ID.String() > lastNonEmpty.ID.String()) {
				lastNonEmpty = m
			}
		}
	}
	if lastActivity == nil {
		return "", sessionCreatedAt
	}
	updatedAt = lastActivity.CreatedAt
	if lastNonEmpty == nil {
		return "", updatedAt
	}
	preview := strings.TrimSpace(lastNonEmpty.Content)
	if len(preview) > 80 {
		preview = preview[:80]
	}
	return preview, updatedAt
}

// checkQuota returns nil if the user is allowed, or a quota error.
func (uc *UseCase) checkQuota(ctx context.Context, userID string) error {
	q, err := uc.quotaRepo.GetUserQuota(ctx, userID)
	if err != nil {
		observability.Warn("quota.fetch_failed", map[string]any{"user_id": userID, "err": err.Error()})
		return nil
	}
	if q.QuotaBypass {
		return nil
	}
	usage, err := uc.quotaRepo.GetUserTokenUsage(ctx, userID)
	if err != nil {
		observability.Warn("quota.usage_fetch_failed", map[string]any{"user_id": userID, "err": err.Error()})
		return nil
	}
	if q.DailyTokenLimit > 0 && usage.DailyTotal >= q.DailyTokenLimit {
		return domain.ErrQuotaDailyExceeded
	}
	if q.WeeklyTokenLimit > 0 && usage.WeeklyTotal >= q.WeeklyTokenLimit {
		return domain.ErrQuotaWeeklyExceeded
	}
	return nil
}

// MeUsage loads quota limits and token usage in parallel (two independent store calls).
func (uc *UseCase) MeUsage(ctx context.Context, userID string) (domain.MeUsage, error) {
	var (
		q    domain.UserQuota
		u    domain.UserTokenUsage
		errQ error
		errU error
		wg   sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		q, errQ = uc.quotaRepo.GetUserQuota(ctx, userID)
	}()
	go func() {
		defer wg.Done()
		u, errU = uc.quotaRepo.GetUserTokenUsage(ctx, userID)
	}()
	wg.Wait()
	if errQ != nil {
		return domain.MeUsage{}, errQ
	}
	if errU != nil {
		return domain.MeUsage{}, errU
	}
	return domain.MeUsage{
		QuotaBypass: q.QuotaBypass,
		Daily: domain.MeUsageWindow{
			LimitTokens: q.DailyTokenLimit,
			UsedTokens:  u.DailyTotal,
		},
		Weekly: domain.MeUsageWindow{
			LimitTokens: q.WeeklyTokenLimit,
			UsedTokens:  u.WeeklyTotal,
		},
	}, nil
}

func (uc *UseCase) SearchChats(
	ctx context.Context,
	userID, query string,
	limit int,
	cursorToken string,
) (SearchResult, error) {
	q := strings.TrimSpace(query)
	if len(q) < 2 {
		return SearchResult{}, domain.ErrSearchQueryTooShort
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	cursor, err := decodeSearchCursor(cursorToken)
	if err != nil {
		return SearchResult{}, err
	}

	raw, err := uc.repo.SearchChats(ctx, domain.SearchChatParams{
		UserID: userID,
		Query:  q,
		Limit:  limit + 1, // one extra row for hasNext
		Cursor: cursor,
	})
	if err != nil {
		return SearchResult{}, err
	}

	hasNext := len(raw.Items) > limit
	items := raw.Items
	if hasNext {
		items = items[:limit]
	}

	next := ""
	if hasNext && len(items) > 0 {
		last := items[len(items)-1]
		next = encodeSearchCursor(domain.SearchCursor{
			SortAt:    last.SortAt,
			SessionID: last.SessionID,
		})
	}

	return SearchResult{
		TotalCount: raw.TotalCount,
		Items:      items,
		HasNext:    hasNext,
		NextCursor: next,
	}, nil
}

type searchCursorWire struct {
	SortAt    string `json:"sortAt"`
	SessionID string `json:"sessionId"`
}

func encodeSearchCursor(c domain.SearchCursor) string {
	wire := searchCursorWire{
		SortAt:    c.SortAt.UTC().Format(time.RFC3339Nano),
		SessionID: c.SessionID,
	}
	b, _ := json.Marshal(wire)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeSearchCursor(token string) (*domain.SearchCursor, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(t)
	if err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	var wire searchCursorWire
	if err := json.Unmarshal(b, &wire); err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	if strings.TrimSpace(wire.SessionID) == "" {
		return nil, domain.ErrInvalidSearchCursor
	}
	sortAt, err := time.Parse(time.RFC3339Nano, wire.SortAt)
	if err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	return &domain.SearchCursor{
		SortAt:    sortAt.UTC(),
		SessionID: wire.SessionID,
	}, nil
}

func (uc *UseCase) GetChatMessagesPage(
	ctx context.Context,
	userID, chatID string,
	limit int,
	direction string,
	cursorToken string,
) (ChatMessagesPage, error) {
	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return ChatMessagesPage{}, err
	}
	if session.UserID != userID {
		return ChatMessagesPage{}, domain.ErrUnauthorized
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	if direction == "" {
		direction = "older"
	}
	if direction != "older" && direction != "newer" {
		return ChatMessagesPage{}, domain.ErrInvalidDirection
	}
	cursor, err := decodeMessageCursor(cursorToken)
	if err != nil {
		return ChatMessagesPage{}, err
	}
	rows, total, err := uc.repo.GetMessagesBySessionPage(ctx, chatID, limit+1, direction, cursor)
	if err != nil {
		return ChatMessagesPage{}, err
	}
	hasNext := len(rows) > limit
	if hasNext {
		rows = rows[:limit]
	}
	// Keep UI order stable (oldest -> newest) for both directions.
	if direction == "older" {
		slices.Reverse(rows)
	}
	if err := uc.attachMessageLiked(ctx, userID, rows); err != nil {
		return ChatMessagesPage{}, err
	}

	next := ""
	if hasNext && len(rows) > 0 {
		oldest := rows[0]
		next = encodeMessageCursor(domain.MessageCursor{
			CreatedAt: oldest.CreatedAt,
			MessageID: oldest.ID.String(),
		})
	}
	return ChatMessagesPage{
		TotalCount: total,
		Messages:   rows,
		HasNext:    hasNext,
		NextCursor: next,
		Direction:  direction,
	}, nil
}

type messageCursorWire struct {
	CreatedAt string `json:"createdAt"`
	MessageID string `json:"messageId"`
}

func encodeMessageCursor(c domain.MessageCursor) string {
	wire := messageCursorWire{
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339Nano),
		MessageID: c.MessageID,
	}
	b, _ := json.Marshal(wire)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeMessageCursor(token string) (*domain.MessageCursor, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(t)
	if err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	var wire messageCursorWire
	if err := json.Unmarshal(b, &wire); err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, wire.CreatedAt)
	if err != nil || strings.TrimSpace(wire.MessageID) == "" {
		return nil, domain.ErrInvalidSearchCursor
	}
	return &domain.MessageCursor{
		CreatedAt: createdAt.UTC(),
		MessageID: wire.MessageID,
	}, nil
}

type sessionCursorWire struct {
	SortAt    string `json:"sortAt"`
	SessionID string `json:"sessionId"`
}

type SessionListResult struct {
	TotalCount int
	Items      []domain.SessionListItem
	HasNext    bool
	NextCursor string
}

func (uc *UseCase) ListSessionsPage(ctx context.Context, userID string, limit int, cursorToken string) (SessionListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	cursor, err := decodeSessionCursor(cursorToken)
	if err != nil {
		return SessionListResult{}, err
	}
	page, err := uc.repo.GetChatSessionsByUserPage(ctx, userID, limit+1, cursor)
	if err != nil {
		return SessionListResult{}, err
	}
	hasNext := len(page.Items) > limit
	items := page.Items
	if hasNext {
		items = items[:limit]
	}
	next := ""
	if hasNext && len(items) > 0 {
		last := items[len(items)-1]
		next = encodeSessionCursor(domain.SessionCursor{
			SortAt:    last.SortAt,
			SessionID: last.Session.ID.String(),
		})
	}
	return SessionListResult{
		TotalCount: page.TotalCount,
		Items:      items,
		HasNext:    hasNext,
		NextCursor: next,
	}, nil
}

func encodeSessionCursor(c domain.SessionCursor) string {
	wire := sessionCursorWire{
		SortAt:    c.SortAt.UTC().Format(time.RFC3339Nano),
		SessionID: c.SessionID,
	}
	b, _ := json.Marshal(wire)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeSessionCursor(token string) (*domain.SessionCursor, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(t)
	if err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	var wire sessionCursorWire
	if err := json.Unmarshal(b, &wire); err != nil {
		return nil, domain.ErrInvalidSearchCursor
	}
	sortAt, err := time.Parse(time.RFC3339Nano, wire.SortAt)
	if err != nil || strings.TrimSpace(wire.SessionID) == "" {
		return nil, domain.ErrInvalidSearchCursor
	}
	return &domain.SessionCursor{
		SortAt:    sortAt.UTC(),
		SessionID: wire.SessionID,
	}, nil
}

func (uc *UseCase) SendMessage(
	ctx context.Context,
	userID, chatID, providerName, model, content string,
	messages []domain.ChatMessage,
) (domain.ChatMessage, map[string]any, error) {
	if strings.TrimSpace(content) == "" {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == domain.RoleUser && strings.TrimSpace(messages[i].Content) != "" {
				content = messages[i].Content
				break
			}
		}
	}
	if strings.TrimSpace(content) == "" {
		return domain.ChatMessage{}, nil, domain.ErrMissingContent
	}

	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	if session.UserID != userID {
		return domain.ChatMessage{}, nil, domain.ErrUnauthorized
	}

	if err := uc.checkQuota(ctx, userID); err != nil {
		return domain.ChatMessage{}, nil, err
	}

	p, err := uc.registry.Get(providerName)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	resolvedProvider := p.Name()
	effModel := effectiveLLMModel(uc.registry, resolvedProvider, model, "")
	if err := uc.ensureModelActive(ctx, resolvedProvider, effModel); err != nil {
		return domain.ChatMessage{}, nil, err
	}

	// Save user message first so trigger creates pending audit row.
	userMsg, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, content, resolvedProvider, effModel)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}

	history, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}

	observability.LLMRequest(resolvedProvider, model, userID, chatID)
	start := time.Now()

	resp, llmErr := p.Complete(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: history,
	})
	if llmErr != nil {
		if ctx.Err() == context.Canceled {
			observability.LLMCancelled(resolvedProvider, model, userID, chatID, 0)
			uc.auditCancel(userMsg.ID.String())
			return domain.ChatMessage{}, nil, domain.ErrUserCancelled
		}
		observability.LLMError(resolvedProvider, model, llmErr)
		uc.auditFail(ctx, userMsg.ID.String(), llmErr)
		return domain.ChatMessage{}, nil, llmErr
	}

	usage := domain.NormalizeUsage(resp.Usage)
	observability.LLMResponse(resolvedProvider, usage.Model, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, time.Since(start).Milliseconds())

	if usage.Model != "" {
		effModel = usage.Model
	}

	assistant, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, resp.Content, resolvedProvider, effModel)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
		observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
	}

	uc.auditUsage(ctx, userMsg.ID.String(), usage)

	assistantOut := []domain.ChatMessage{assistant}
	if err := uc.attachMessageLiked(ctx, userID, assistantOut); err != nil {
		return domain.ChatMessage{}, nil, err
	}
	return assistantOut[0], resp.Usage, nil
}

func (uc *UseCase) SyncMessages(
	ctx context.Context,
	userID, chatID, providerName, model string,
	pending []domain.BatchMessage,
) (SyncResult, error) {
	if len(pending) == 0 {
		return SyncResult{}, domain.ErrMissingContent
	}

	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return SyncResult{}, err
	}
	if session.UserID != userID {
		return SyncResult{}, domain.ErrUnauthorized
	}

	if err := uc.checkQuota(ctx, userID); err != nil {
		return SyncResult{}, err
	}

	p, err := uc.registry.Get(providerName)
	if err != nil {
		return SyncResult{}, err
	}
	resolvedProvider := p.Name()
	effModel := effectiveLLMModel(uc.registry, resolvedProvider, model, "")
	if err := uc.ensureModelActive(ctx, resolvedProvider, effModel); err != nil {
		return SyncResult{}, err
	}

	prepared := make([]domain.BatchMessage, 0, len(pending))
	for _, msg := range pending {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		prepared = append(prepared, domain.BatchMessage{
			Content:  content,
			Provider: resolvedProvider,
			Model:    effModel,
		})
	}
	if len(prepared) == 0 {
		return SyncResult{}, domain.ErrMissingContent
	}

	savedUsers, err := uc.repo.SaveMessages(ctx, chatID, userID, prepared)
	if err != nil {
		return SyncResult{}, err
	}

	history, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		return SyncResult{}, err
	}

	observability.LLMRequest(resolvedProvider, model, userID, chatID)
	start := time.Now()
	resp, llmErr := p.Complete(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: history,
	})
	if llmErr != nil {
		observability.LLMError(resolvedProvider, model, llmErr)
		if len(savedUsers) > 0 {
			uc.auditFail(ctx, savedUsers[len(savedUsers)-1].ID.String(), llmErr)
		}
		return SyncResult{}, llmErr
	}

	usage := domain.NormalizeUsage(resp.Usage)
	observability.LLMResponse(resolvedProvider, usage.Model, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, time.Since(start).Milliseconds())
	if usage.Model != "" {
		effModel = usage.Model
	}

	assistant, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, resp.Content, resolvedProvider, effModel)
	if err != nil {
		if len(savedUsers) > 0 {
			uc.auditFail(ctx, savedUsers[len(savedUsers)-1].ID.String(), err)
		}
		return SyncResult{}, err
	}
	if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
		observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
	}
	if len(savedUsers) > 0 {
		uc.auditUsage(ctx, savedUsers[len(savedUsers)-1].ID.String(), usage)
	}

	if err := uc.attachMessageLiked(ctx, userID, savedUsers); err != nil {
		return SyncResult{}, err
	}
	assistantOut := []domain.ChatMessage{assistant}
	if err := uc.attachMessageLiked(ctx, userID, assistantOut); err != nil {
		return SyncResult{}, err
	}

	return SyncResult{
		SyncedMessages:   savedUsers,
		AssistantMessage: assistantOut[0],
		Usage:            resp.Usage,
	}, nil
}

func (uc *UseCase) StreamMessage(
	ctx context.Context,
	userID, chatID, providerName, model, content string,
	messages []domain.ChatMessage,
) (<-chan domain.StreamEvent, map[string]any, StreamFinalize, StreamCancel, error) {
	prompt := strings.TrimSpace(content)
	if prompt == "" {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == domain.RoleUser && strings.TrimSpace(messages[i].Content) != "" {
				prompt = messages[i].Content
				break
			}
		}
	}
	if prompt == "" {
		return nil, nil, nil, nil, domain.ErrMissingContent
	}

	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if session.UserID != userID {
		return nil, nil, nil, nil, domain.ErrUnauthorized
	}

	if err := uc.checkQuota(ctx, userID); err != nil {
		return nil, nil, nil, nil, err
	}

	p, err := uc.registry.Get(providerName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	resolvedProvider := p.Name()
	effModel := effectiveLLMModel(uc.registry, resolvedProvider, model, "")
	if err := uc.ensureModelActive(ctx, resolvedProvider, effModel); err != nil {
		return nil, nil, nil, nil, err
	}

	// Save user message before stream so trigger creates pending audit row.
	userMsg, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, prompt, resolvedProvider, effModel)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	userMsgID := userMsg.ID.String()

	history, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		uc.auditFail(ctx, userMsgID, err)
		return nil, nil, nil, nil, err
	}

	assistantID := uuid.New()
	if _, err := uc.repo.SaveAssistantPlaceholder(ctx, chatID, assistantID.String(), resolvedProvider, effModel); err != nil {
		uc.auditFail(ctx, userMsgID, err)
		return nil, nil, nil, nil, err
	}
	assistantMsgID := assistantID.String()

	observability.LLMRequest(resolvedProvider, model, userID, chatID)

	events, err := p.Stream(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: history,
	})
	if err != nil {
		observability.LLMError(resolvedProvider, model, err)
		_ = uc.repo.SoftDeleteChatMessageByID(ctx, chatID, assistantMsgID)
		uc.auditFail(ctx, userMsgID, err)
		return nil, nil, nil, nil, err
	}

	usageMeta := map[string]any{
		"provider":           resolvedProvider,
		"model":              effModel,
		"userMessageId":      userMsgID,
		"assistantMessageId": assistantMsgID,
	}
	finalize := func(assistantContent string, streamUsage map[string]any) (domain.ChatMessage, error) {
		observability.Info("llm.stream.complete", map[string]any{
			"provider":   resolvedProvider,
			"model":      effModel,
			"session_id": chatID,
			"chars":      len(assistantContent),
		})
		trimmed := strings.TrimSpace(assistantContent)
		if trimmed == "" {
			_ = uc.repo.SoftDeleteChatMessageByID(ctx, chatID, assistantMsgID)
			uc.auditFail(ctx, userMsgID, errors.New("empty assistant response"))
			return domain.ChatMessage{}, domain.ErrMissingContent
		}
		msg, err := uc.repo.UpdateAssistantMessageContent(ctx, chatID, assistantMsgID, assistantContent, resolvedProvider, effModel)
		if err != nil {
			_ = uc.repo.SoftDeleteChatMessageByID(ctx, chatID, assistantMsgID)
			uc.auditFail(ctx, userMsgID, err)
			return domain.ChatMessage{}, err
		}
		if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
			observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
		}
		usage := mergeStreamUsageForAudit(resolvedProvider, effModel, streamUsage)
		if usage.TotalTokens > 0 || usage.PromptTokens > 0 || usage.CompletionTokens > 0 {
			uc.auditUsage(ctx, userMsgID, usage)
		}
		return msg, nil
	}
	cancel := func(partialChars int) {
		observability.LLMCancelled(resolvedProvider, effModel, userID, chatID, partialChars)
		if partialChars == 0 {
			// Stop / disconnect with nothing to save: roll back this turn (user + empty assistant), önceki mesajlar kalır.
			cleanCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
			_ = uc.repo.SoftDeleteChatMessageByID(cleanCtx, chatID, assistantMsgID)
			_ = uc.repo.SoftDeleteUserMessage(cleanCtx, chatID, userMsgID, userID)
			c()
		}
		uc.auditCancel(userMsgID)
	}

	return events, usageMeta, finalize, cancel, nil
}

func mergeStreamUsageForAudit(resolvedProvider, effModel string, streamUsage map[string]any) domain.NormalizedUsage {
	raw := make(map[string]any)
	for k, v := range streamUsage {
		raw[k] = v
	}
	if !usageMapHasNonEmptyString(raw, "provider") {
		raw["provider"] = resolvedProvider
	}
	if !usageMapHasNonEmptyString(raw, "model") {
		raw["model"] = effModel
	}
	return domain.NormalizeUsage(raw)
}

func usageMapHasNonEmptyString(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) != ""
}

// auditFail records an LLM failure in llm_interaction_log via RPC.
func (uc *UseCase) auditFail(ctx context.Context, userMessageID string, llmErr error) {
	code := domain.LLMErrorCode(llmErr)
	status := domain.LLMHTTPStatus(llmErr)
	summary := llmErr.Error()
	if len(summary) > 500 {
		summary = summary[:500]
	}
	if err := uc.quotaRepo.FailPendingLog(ctx, userMessageID, summary, code, status); err != nil {
		observability.Warn("audit.fail_pending_rpc", map[string]any{"user_message_id": userMessageID, "err": err.Error()})
	}
}

// auditCancel marks the pending interaction log as cancelled.
func (uc *UseCase) auditCancel(userMessageID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := uc.quotaRepo.CancelPendingLog(ctx, userMessageID); err != nil {
		observability.Warn("audit.cancel_pending_rpc", map[string]any{"user_message_id": userMessageID, "err": err.Error()})
	}
}

// auditUsage records token usage in llm_interaction_log via RPC.
func (uc *UseCase) auditUsage(ctx context.Context, userMessageID string, usage domain.NormalizedUsage) {
	if usage.TotalTokens == 0 && usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
		return
	}
	if err := uc.quotaRepo.SetUsageLog(ctx, userMessageID, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens); err != nil {
		observability.Warn("audit.set_usage_rpc", map[string]any{"user_message_id": userMessageID, "err": err.Error()})
	}
}

func effectiveLLMModel(reg *provider.Registry, resolvedProvider, requestModel, usageModel string) string {
	if s := strings.TrimSpace(usageModel); s != "" {
		return s
	}
	if s := strings.TrimSpace(requestModel); s != "" {
		return s
	}
	if meta, ok := reg.Meta(resolvedProvider); ok {
		return strings.TrimSpace(meta.DefaultModel)
	}
	return ""
}

func (uc *UseCase) ListProviders() []provider.ProviderMeta {
	return uc.registry.List()
}

func (uc *UseCase) ListSupportedModels(ctx context.Context) ([]domain.SupportedModel, error) {
	if uc.models == nil {
		return catalog.SupportedModelsFromRegistry(uc.registry), nil
	}
	return uc.models.ListActiveSupportedModels(ctx)
}

func (uc *UseCase) DefaultProvider() string {
	return uc.registry.Default()
}

func (uc *UseCase) ensureModelActive(ctx context.Context, resolvedProvider, effModel string) error {
	if uc.models == nil {
		return nil
	}
	em := strings.TrimSpace(effModel)
	if em == "" {
		return nil
	}
	ok, err := uc.models.IsModelActive(ctx, resolvedProvider, em)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrModelDiscontinued
	}
	return nil
}
