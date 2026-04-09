package chat

import (
	"context"
	"strings"
	"time"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/observability"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
)

type StreamFinalize func(assistantContent string) (domain.ChatMessage, error)
type StreamCancel func(partialChars int)

type SyncResult struct {
	SyncedMessages   []domain.ChatMessage
	AssistantMessage domain.ChatMessage
	Usage            map[string]any
}

type UseCase struct {
	repo      domain.Repository
	quotaRepo domain.QuotaRepository
	registry  *provider.Registry
}

func NewUseCase(repo domain.Repository, quotaRepo domain.QuotaRepository, registry *provider.Registry) *UseCase {
	return &UseCase{repo: repo, quotaRepo: quotaRepo, registry: registry}
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

func (uc *UseCase) GetSessionSummary(ctx context.Context, chatID string) (lastMessagePreview string, updatedAt time.Time) {
	msgs, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil || len(msgs) == 0 {
		return "", time.Now().UTC()
	}
	last := msgs[len(msgs)-1]
	preview := strings.TrimSpace(last.Content)
	if len(preview) > 80 {
		preview = preview[:80]
	}
	return preview, last.CreatedAt
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

	return assistant, resp.Usage, nil
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

	return SyncResult{
		SyncedMessages:   savedUsers,
		AssistantMessage: assistant,
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

	// Save user message before stream so trigger creates pending audit row.
	userMsg, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, prompt, resolvedProvider, effModel)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	observability.LLMRequest(resolvedProvider, model, userID, chatID)

	events, err := p.Stream(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: func() []domain.ChatMessage {
			h, _ := uc.repo.GetMessagesBySession(ctx, chatID)
			return h
		}(),
	})
	if err != nil {
		observability.LLMError(resolvedProvider, model, err)
		uc.auditFail(ctx, userMsg.ID.String(), err)
		return nil, nil, nil, nil, err
	}

	usageMeta := map[string]any{
		"provider": resolvedProvider,
		"model":    effModel,
	}

	userMsgID := userMsg.ID.String()
	finalize := func(assistantContent string) (domain.ChatMessage, error) {
		observability.Info("llm.stream.complete", map[string]any{
			"provider":   resolvedProvider,
			"model":      effModel,
			"session_id": chatID,
			"chars":      len(assistantContent),
		})
		msg, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, assistantContent, resolvedProvider, effModel)
		if err != nil {
			uc.auditFail(ctx, userMsgID, err)
			return domain.ChatMessage{}, err
		}
		if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
			observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
		}
		return msg, nil
	}
	cancel := func(partialChars int) {
		observability.LLMCancelled(resolvedProvider, effModel, userID, chatID, partialChars)
		uc.auditCancel(userMsgID)
	}

	return events, usageMeta, finalize, cancel, nil
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
	if usage.TotalTokens == 0 && usage.PromptTokens == 0 {
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

func (uc *UseCase) DefaultProvider() string {
	return uc.registry.Default()
}
