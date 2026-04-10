package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps go-redis with prefix delete via SCAN.
type Redis struct {
	c   *redis.Client
	ttl time.Duration // optional max TTL cap for Set (0 = use passed ttl only)
}

func NewRedis(addr, password string, db int) *Redis {
	return &Redis{
		c: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, bool) {
	s, err := r.c.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		return nil, false
	}
	return []byte(s), true
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = r.ttl
	}
	_ = r.c.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) DeletePrefix(ctx context.Context, prefix string) {
	if prefix == "" {
		return
	}
	iter := r.c.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		_ = r.c.Del(ctx, iter.Val()).Err()
	}
	_ = iter.Err()
}
