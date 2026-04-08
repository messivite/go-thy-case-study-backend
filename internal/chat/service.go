package chat

import (
	"context"
	"fmt"

	"github.com/example/thy-case-study-backend/internal/provider"
	"github.com/example/thy-case-study-backend/internal/repo"
)

type ChatService struct {
	repo            repo.Repository
	providerFactory *provider.ProviderFactory
}

func NewChatService(repository repo.Repository, providerFactory *provider.ProviderFactory) *ChatService {
	return &ChatService{repo: repository, providerFactory: providerFactory}
}

func (s *ChatService) CreateSession(ctx context.Context, userID string, title string) (repo.ChatSession, error) {
	return s.repo.CreateChatSession(ctx, userID, title)
}

func (s *ChatService) ListSessions(ctx context.Context, userID string) ([]repo.ChatSession, error) {
	return s.repo.GetChatSessionsByUser(ctx, userID)
}

func (s *ChatService) GetSessionMessages(ctx context.Context, userID string, sessionID string) ([]repo.ChatMessage, error) {
	session, err := s.repo.GetChatSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, fmt.Errorf("not authorized to view session")
	}
	return s.repo.GetMessagesBySession(ctx, sessionID)
}

func (s *ChatService) SendMessage(ctx context.Context, userID string, sessionID string, providerName string, content string) (repo.ChatMessage, repo.ChatMessage, error) {
	session, err := s.repo.GetChatSessionByID(ctx, sessionID)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}
	if session.UserID != userID {
		return repo.ChatMessage{}, repo.ChatMessage{}, fmt.Errorf("not authorized to post messages")
	}

	userMessage, err := s.repo.SaveMessage(ctx, sessionID, userID, "user", content)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}

	history, err := s.repo.GetMessagesBySession(ctx, sessionID)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}

	provider, err := s.providerFactory.GetProvider(providerName)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}

	assistantContent, err := provider.Respond(ctx, session, history, content)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}

	assistantMessage, err := s.repo.SaveMessage(ctx, sessionID, "", "assistant", assistantContent)
	if err != nil {
		return repo.ChatMessage{}, repo.ChatMessage{}, err
	}

	return userMessage, assistantMessage, nil
}
