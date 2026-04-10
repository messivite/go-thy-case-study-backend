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

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

const anthropicAPIVersion = "2023-06-01"

type AnthropicProvider struct {
	apiKey string
	model  string
	name   string // registry / providers.yaml adı (anthropic veya claude)
	client *http.Client
}

// NewAnthropicProvider Claude Messages API; registry adı "anthropic".
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return NewAnthropicProviderNamed(apiKey, model, "anthropic")
}

// NewAnthropicProviderNamed aynı API; providers.yaml'da name: claude kullanılırsa registryName "claude" ver.
func NewAnthropicProviderNamed(apiKey, model, registryName string) *AnthropicProvider {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	if registryName == "" {
		registryName = "anthropic"
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		name:   registryName,
		client: NewHTTPClient(DefaultClientConfig),
	}
}

func (p *AnthropicProvider) Name() string { return p.name }

func (p *AnthropicProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	if p.apiKey == "" {
		return domain.ProviderResponse{}, fmt.Errorf("%w: ANTHROPIC_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	sys, msgs := toAnthropicMessages(req.Messages)
	body := anthropicRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages:  msgs,
		System:    sys,
		Stream:    false,
	}

	respBody, err := p.doRequest(ctx, body)
	if err != nil {
		return domain.ProviderResponse{}, err
	}
	defer respBody.Close()

	var resp anthropicMessageResponse
	if err := json.NewDecoder(respBody).Decode(&resp); err != nil {
		return domain.ProviderResponse{}, fmt.Errorf("anthropic response parse: %w", err)
	}

	content := extractAnthropicTextContent(resp.Content)

	usage := map[string]any{
		"provider":          "anthropic",
		"model":             resp.Model,
		"prompt_tokens":     resp.Usage.InputTokens,
		"completion_tokens": resp.Usage.OutputTokens,
		"total_tokens":      resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	return domain.ProviderResponse{Content: content, Usage: usage}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: ANTHROPIC_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	sys, msgs := toAnthropicMessages(req.Messages)
	body := anthropicRequest{
		Model:     model,
		MaxTokens: 8192,
		Messages:  msgs,
		System:    sys,
		Stream:    true,
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
		const maxLine = 1024 * 1024
		scanner.Buffer(make([]byte, 0, 64*1024), maxLine)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if data == "" || data == "[DONE]" {
				continue
			}

			var raw map[string]any
			if err := json.Unmarshal([]byte(data), &raw); err != nil {
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventError, Message: "anthropic stream parse: " + err.Error()}:
				}
				return
			}

			typ, _ := raw["type"].(string)
			switch typ {
			case "content_block_delta":
				delta, _ := raw["delta"].(map[string]any)
				if delta == nil {
					continue
				}
				dt, _ := delta["type"].(string)
				if dt != "text_delta" {
					continue
				}
				text, _ := delta["text"].(string)
				if text == "" {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventDelta, Delta: text}:
				}
			case "message_delta":
				// usage burada da gelebilir; şimdilik yok sayıyoruz (finalize repo tarafında)
			case "message_stop":
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventDone}:
				}
				return
			case "error":
				errObj, _ := raw["error"].(map[string]any)
				msg := "anthropic error"
				if errObj != nil {
					if m, ok := errObj["message"].(string); ok && m != "" {
						msg = m
					}
				}
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventError, Message: msg}:
				}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
				return
			case events <- domain.StreamEvent{Type: domain.EventError, Message: err.Error()}:
			}
			return
		}
		select {
		case <-ctx.Done():
			return
		case events <- domain.StreamEvent{Type: domain.EventDone}:
		}
	}()

	return events, nil
}

func (p *AnthropicProvider) doRequest(ctx context.Context, body anthropicRequest) (io.ReadCloser, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	cfg := DefaultClientConfig
	if body.Stream {
		cfg = StreamClientConfig
	}

	resp, err := DoWithRetry(ctx, p.client, httpReq, cfg, "anthropic")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		mappedErr := MapHTTPError(resp.StatusCode, "anthropic")
		return nil, fmt.Errorf("%w — %s", mappedErr, string(errBody))
	}

	return resp.Body, nil
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessageResponse struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func extractAnthropicTextContent(blocks []struct {
	Type string `json:"type"`
	Text string `json:"text"`
}) string {
	var b strings.Builder
	for _, c := range blocks {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	return b.String()
}

// toAnthropicMessages: system ayrı; user/assistant sırası API kuralına uygun (aynı rol ardışık satırlar birleştirilir).
func toAnthropicMessages(messages []domain.ChatMessage) (system string, out []anthropicMessage) {
	var sysParts []string
	for _, m := range messages {
		if m.Role == domain.RoleSystem {
			if strings.TrimSpace(m.Content) != "" {
				sysParts = append(sysParts, strings.TrimSpace(m.Content))
			}
			continue
		}
		role := "user"
		if m.Role == domain.RoleAssistant {
			role = "assistant"
		}
		if m.Role != domain.RoleUser && m.Role != domain.RoleAssistant {
			continue
		}
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Role == role {
			out[len(out)-1].Content += "\n\n" + content
		} else {
			out = append(out, anthropicMessage{Role: role, Content: content})
		}
	}
	if len(sysParts) > 0 {
		system = strings.Join(sysParts, "\n\n")
	}
	return system, out
}
