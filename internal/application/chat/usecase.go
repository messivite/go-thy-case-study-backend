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

func (uc *UseCase) CreateSession(ctx context.Context, userID, title string) (domain.ChatSession, error) {
	return uc.repo.CreateChatSession(ctx, userID, title)
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

	observability.LLMRequest(providerName, model, userID, chatID)
	start := time.Now()

	resp, err := p.Complete(ctx, domain.ProviderRequest{
		Provider: providerName,
		Model:    model,
		Messages: reqMessages,
	})
	if err != nil {
		observability.LLMError(providerName, model, err)
		return domain.ChatMessage{}, nil, err
	}

	usage := domain.NormalizeUsage(resp.Usage)
	observability.LLMResponse(providerName, usage.Model, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, time.Since(start).Milliseconds())

	if _, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, content); err != nil {
		return domain.ChatMessage{}, nil, err
	}

	assistant, err := uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, resp.Content)
	if err != nil {
		return domain.ChatMessage{}, nil, err
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

	observability.LLMRequest(providerName, model, userID, chatID)

	events, err := p.Stream(ctx, domain.ProviderRequest{
		Provider: providerName,
		Model:    model,
		Messages: reqMessages,
	})
	if err != nil {
		observability.LLMError(providerName, model, err)
		return nil, nil, nil, err
	}

	if _, err := uc.repo.SaveMessage(ctx, chatID, userID, domain.RoleUser, prompt); err != nil {
		return nil, nil, nil, err
	}

	usage := map[string]any{
		"provider": providerName,
		"model":    model,
	}

	finalize := func(assistantContent string) (domain.ChatMessage, error) {
		observability.Info("llm.stream.complete", map[string]any{
			"provider":   providerName,
			"model":      model,
			"session_id": chatID,
			"chars":      len(assistantContent),
		})
		return uc.repo.SaveMessage(ctx, chatID, "", domain.RoleAssistant, assistantContent)
	}

	return events, usage, finalize, nil
}

func (uc *UseCase) ListProviders() []provider.ProviderMeta {
	return uc.registry.List()
}

func (uc *UseCase) DefaultProvider() string {
	return uc.registry.Default()
}
