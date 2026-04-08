package chat

import "context"

type LLMProvider interface {
	Name() string
	Complete(ctx context.Context, req ProviderRequest) (ProviderResponse, error)
	Stream(ctx context.Context, req ProviderRequest) (<-chan StreamEvent, error)
}
