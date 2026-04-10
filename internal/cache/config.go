package cache

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// FromEnv builds a Store: disabled -> Nop; backend memory|redis.
func FromEnv() (Store, time.Duration, time.Duration) {
	if !truthy(os.Getenv("CACHE_ENABLED")) {
		return Nop{}, 0, 0
	}
	ttlList := durationSec("CACHE_TTL_CHAT_LIST_SEC", 20*time.Second)
	ttlMsgs := durationSec("CACHE_TTL_CHAT_MESSAGES_SEC", 15*time.Second)

	backend := strings.ToLower(strings.TrimSpace(envOr("CACHE_BACKEND", "memory")))
	switch backend {
	case "redis":
		addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		db, _ := strconv.Atoi(strings.TrimSpace(envOr("REDIS_DB", "0")))
		pw := os.Getenv("REDIS_PASSWORD")
		return NewRedis(addr, pw, db), ttlList, ttlMsgs
	default:
		return NewMemory(), ttlList, ttlMsgs
	}
}

func truthy(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func durationSec(env string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(env))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return time.Duration(n) * time.Second
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
