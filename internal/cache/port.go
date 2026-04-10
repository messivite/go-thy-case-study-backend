package cache

import (
	"context"
	"time"
)

// Store abstracts HTTP response body caching (JSON bytes) with prefix invalidation.
type Store interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	DeletePrefix(ctx context.Context, prefix string)
}

// Nop is a no-op implementation (always miss, no writes).
type Nop struct{}

func (Nop) Get(context.Context, string) ([]byte, bool) { return nil, false }

func (Nop) Set(context.Context, string, []byte, time.Duration) {}

func (Nop) DeletePrefix(context.Context, string) {}
