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

type UseCase struct {
	repo     domain.Repository
	registry *provider.Registry
}

func NewUseCase(repo domain.Repository, registry *provider.Registry) *UseCase {
	return &UseCase{repo: repo, registry: registry}
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

	history, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	reqMessages := append(history, domain.ChatMessage{
		Role:    domain.RoleUser,
		Content: content,
	})

	p, err := uc.registry.Get(providerName)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	resolvedProvider := p.Name()

	observability.LLMRequest(resolvedProvider, model, userID, chatID)
	start := time.Now()

	resp, err := p.Complete(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: reqMessages,
	})
	if err != nil {
		observability.LLMError(resolvedProvider, model, err)
		return domain.ChatMessage{}, nil, err
	}

	usage := domain.NormalizeUsage(resp.Usage)
	observability.LLMResponse(resolvedProvider, usage.Model, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, time.Since(start).Milliseconds())

	effModel := effectiveLLMModel(uc.registry, resolvedProvider, model, usage.Model)

	if _, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, content, resolvedProvider, effModel); err != nil {
		return domain.ChatMessage{}, nil, err
	}

	assistant, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, resp.Content, resolvedProvider, effModel)
	if err != nil {
		return domain.ChatMessage{}, nil, err
	}
	if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
		observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
	}

	return assistant, resp.Usage, nil
}

func (uc *UseCase) StreamMessage(
	ctx context.Context,
	userID, chatID, providerName, model, content string,
	messages []domain.ChatMessage,
) (<-chan domain.StreamEvent, map[string]any, StreamFinalize, error) {
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
		return nil, nil, nil, domain.ErrMissingContent
	}

	session, err := uc.repo.GetChatSessionByID(ctx, chatID)
	if err != nil {
		return nil, nil, nil, err
	}
	if session.UserID != userID {
		return nil, nil, nil, domain.ErrUnauthorized
	}

	history, err := uc.repo.GetMessagesBySession(ctx, chatID)
	if err != nil {
		return nil, nil, nil, err
	}
	reqMessages := append(history, domain.ChatMessage{
		Role:    domain.RoleUser,
		Content: prompt,
	})

	p, err := uc.registry.Get(providerName)
	if err != nil {
		return nil, nil, nil, err
	}
	resolvedProvider := p.Name()

	observability.LLMRequest(resolvedProvider, model, userID, chatID)

	events, err := p.Stream(ctx, domain.ProviderRequest{
		Provider: resolvedProvider,
		Model:    model,
		Messages: reqMessages,
	})
	if err != nil {
		observability.LLMError(resolvedProvider, model, err)
		return nil, nil, nil, err
	}

	effModel := effectiveLLMModel(uc.registry, resolvedProvider, model, "")
	if _, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, prompt, resolvedProvider, effModel); err != nil {
		return nil, nil, nil, err
	}

	usage := map[string]any{
		"provider": resolvedProvider,
		"model":    effModel,
	}

	finalize := func(assistantContent string) (domain.ChatMessage, error) {
		observability.Info("llm.stream.complete", map[string]any{
			"provider":   resolvedProvider,
			"model":      effModel,
			"session_id": chatID,
			"chars":      len(assistantContent),
		})
		msg, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, assistantContent, resolvedProvider, effModel)
		if err != nil {
			return domain.ChatMessage{}, err
		}
		if err := uc.repo.UpdateSessionLastLLM(ctx, chatID, resolvedProvider, effModel); err != nil {
			observability.Info("session.last_llm.update_failed", map[string]any{"session_id": chatID, "err": err.Error()})
		}
		return msg, nil
	}

	return events, usage, finalize, nil
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
