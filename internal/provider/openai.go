package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		client: NewHTTPClient(DefaultClientConfig),
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	if p.apiKey == "" {
		return domain.ProviderResponse{}, fmt.Errorf("%w: OPENAI_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	body := openaiRequest{
		Model:    model,
		Messages: toOpenAIMessages(req.Messages),
		Stream:   false,
	}

	respBody, err := p.doRequest(ctx, body)
	if err != nil {
		return domain.ProviderResponse{}, err
	}
	defer respBody.Close()

	var resp openaiChatResponse
	if err := json.NewDecoder(respBody).Decode(&resp); err != nil {
		return domain.ProviderResponse{}, fmt.Errorf("openai response parse hatası: %w", err)
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	usage := map[string]any{
		"provider":          "openai",
		"model":             resp.Model,
		"prompt_tokens":     resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_tokens":      resp.Usage.TotalTokens,
	}

	return domain.ProviderResponse{Content: content, Usage: usage}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: OPENAI_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	body := openaiRequest{
		Model:    model,
		Messages: toOpenAIMessages(req.Messages),
		Stream:   true,
	}

	respBody, err := p.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}

	events := make(chan domain.StreamEvent, 16)

	go func() {
		defer close(events)
		defer respBody.Close()

		scanner := bufio.NewScanner(respBody)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				events <- domain.StreamEvent{Type: domain.EventDone}
				return
			}

			var chunk openaiStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				events <- domain.StreamEvent{Type: domain.EventError, Message: "parse hatası: " + err.Error()}
				return
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventDelta, Delta: chunk.Choices[0].Delta.Content}:
				}
			}
		}

		if err := scanner.Err(); err != nil {
			events <- domain.StreamEvent{Type: domain.EventError, Message: err.Error()}
		}
	}()

	return events, nil
}

func (p *OpenAIProvider) doRequest(ctx context.Context, body openaiRequest) (io.ReadCloser, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai request marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("openai request oluşturma: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	cfg := DefaultClientConfig
	if body.Stream {
		cfg = StreamClientConfig
	}

	resp, err := DoWithRetry(ctx, p.client, httpReq, cfg, "openai")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		mappedErr := MapHTTPError(resp.StatusCode, "openai")
		return nil, fmt.Errorf("%w — %s", mappedErr, string(errBody))
	}

	return resp.Body, nil
}

// ---------------------------------------------------------------------------
// OpenAI API types
// ---------------------------------------------------------------------------

type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openaiStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func toOpenAIMessages(messages []domain.ChatMessage) []openaiMessage {
	out := make([]openaiMessage, 0, len(messages))
	for _, m := range messages {
		out = append(out, openaiMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return out
}
