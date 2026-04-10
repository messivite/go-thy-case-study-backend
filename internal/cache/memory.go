package cache

import (
	"context"
	"sync"
	"time"
)

type memEntry struct {
	value  []byte
	expiry time.Time
}

// Memory is a process-local TTL cache with prefix deletion.
type Memory struct {
	mu   sync.RWMutex
	data map[string]memEntry
}

func NewMemory() *Memory {
	return &Memory{data: make(map[string]memEntry)}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.data[key]
	if !ok {
		return nil, false
	}
	if !e.expiry.IsZero() && time.Now().After(e.expiry) {
		delete(m.data, key)
		return nil, false
	}
	out := make([]byte, len(e.value))
	copy(out, e.value)
	return out, true
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := make([]byte, len(value))
	copy(v, value)
	e := memEntry{value: v}
	if ttl > 0 {
		e.expiry = time.Now().Add(ttl)
	}
	m.data[key] = e
}

func (m *Memory) DeletePrefix(_ context.Context, prefix string) {
	if prefix == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.data, k)
		}
	}
}
