package provider

import domain "github.com/example/thy-case-study-backend/internal/domain/chat"

var _ domain.LLMProvider = (*OpenAIProvider)(nil)
var _ domain.LLMProvider = (*GeminiProvider)(nil)
