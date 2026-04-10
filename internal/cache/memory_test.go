package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemory_SetGetExpire(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	m.Set(ctx, "k", []byte("v"), 50*time.Millisecond)
	b, ok := m.Get(ctx, "k")
	if !ok || string(b) != "v" {
		t.Fatalf("expected hit, got ok=%v b=%q", ok, b)
	}
	time.Sleep(60 * time.Millisecond)
	_, ok = m.Get(ctx, "k")
	if ok {
		t.Fatal("expected miss after TTL")
	}
}

func TestMemory_DeletePrefix(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	m.Set(ctx, "user:a:list:x", []byte("1"), time.Hour)
	m.Set(ctx, "user:a:list:y", []byte("2"), time.Hour)
	m.Set(ctx, "other", []byte("3"), time.Hour)
	m.DeletePrefix(ctx, "user:a:list:")
	if _, ok := m.Get(ctx, "user:a:list:x"); ok {
		t.Fatal("x should be gone")
	}
	if _, ok := m.Get(ctx, "user:a:list:y"); ok {
		t.Fatal("y should be gone")
	}
	b, ok := m.Get(ctx, "other")
	if !ok || string(b) != "3" {
		t.Fatalf("other key intact: ok=%v b=%q", ok, b)
	}
}
