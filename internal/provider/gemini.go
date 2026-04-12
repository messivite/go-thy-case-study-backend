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

type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		client: NewHTTPClient(DefaultClientConfig),
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Complete(ctx context.Context, req domain.ProviderRequest) (domain.ProviderResponse, error) {
	if p.apiKey == "" {
		return domain.ProviderResponse{}, fmt.Errorf("%w: GEMINI_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	body := geminiRequest{
		Contents: toGeminiContents(req.Messages),
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, p.apiKey)

	respBody, err := p.doRequest(ctx, url, body)
	if err != nil {
		return domain.ProviderResponse{}, err
	}
	defer respBody.Close()

	var resp geminiResponse
	if err := json.NewDecoder(respBody).Decode(&resp); err != nil {
		return domain.ProviderResponse{}, fmt.Errorf("gemini response parse hatası: %w", err)
	}

	content := ""
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		content = resp.Candidates[0].Content.Parts[0].Text
	}

	usage := map[string]any{
		"provider":          "gemini",
		"model":             model,
		"prompt_tokens":     resp.UsageMetadata.PromptTokenCount,
		"completion_tokens": resp.UsageMetadata.CandidatesTokenCount,
		"total_tokens":      resp.UsageMetadata.TotalTokenCount,
	}

	return domain.ProviderResponse{Content: content, Usage: usage}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, req domain.ProviderRequest) (<-chan domain.StreamEvent, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("%w: GEMINI_API_KEY boş", domain.ErrProviderAuthFailed)
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	body := geminiRequest{
		Contents: toGeminiContents(req.Messages),
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s&alt=sse", model, p.apiKey)

	respBody, err := p.doRequest(ctx, url, body)
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

			var chunk geminiResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				events <- domain.StreamEvent{Type: domain.EventError, Message: "parse hatası: " + err.Error()}
				return
			}

			um := chunk.UsageMetadata
			if um.PromptTokenCount > 0 || um.CandidatesTokenCount > 0 || um.TotalTokenCount > 0 {
				meta := map[string]any{
					"provider":          "gemini",
					"model":             model,
					"prompt_tokens":     um.PromptTokenCount,
					"completion_tokens": um.CandidatesTokenCount,
					"total_tokens":      um.TotalTokenCount,
				}
				select {
				case <-ctx.Done():
					return
				case events <- domain.StreamEvent{Type: domain.EventMeta, Meta: meta}:
				}
			}

			if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
				text := chunk.Candidates[0].Content.Parts[0].Text
				if text != "" {
					select {
					case <-ctx.Done():
						return
					case events <- domain.StreamEvent{Type: domain.EventDelta, Delta: text}:
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			events <- domain.StreamEvent{Type: domain.EventError, Message: err.Error()}
			return
		}

		events <- domain.StreamEvent{Type: domain.EventDone}
	}()

	return events, nil
}

func (p *GeminiProvider) doRequest(ctx context.Context, url string, body geminiRequest) (io.ReadCloser, error) {
	isStream := strings.Contains(url, "streamGenerateContent")

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini request marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("gemini request oluşturma: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	cfg := DefaultClientConfig
	if isStream {
		cfg = StreamClientConfig
	}

	resp, err := DoWithRetry(ctx, p.client, httpReq, cfg, "gemini")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		mappedErr := MapHTTPError(resp.StatusCode, "gemini")
		return nil, fmt.Errorf("%w — %s", mappedErr, string(errBody))
	}

	return resp.Body, nil
}

// ---------------------------------------------------------------------------
// Gemini API types
// ---------------------------------------------------------------------------

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func toGeminiContents(messages []domain.ChatMessage) []geminiContent {
	out := make([]geminiContent, 0, len(messages))
	for _, m := range messages {
		role := "user"
		if m.Role == domain.RoleAssistant {
			role = "model"
		}
		if m.Role == domain.RoleSystem {
			role = "user"
		}
		out = append(out, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	return out
}
