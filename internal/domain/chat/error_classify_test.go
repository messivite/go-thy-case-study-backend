package chat

import (
	"errors"
	"testing"
)

func TestLLMErrorCode(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{ErrProviderRateLimited, "rate_limited"},
		{ErrProviderUnavailable, "upstream_5xx"},
		{ErrProviderTimeout, "timeout"},
		{ErrProviderAuthFailed, "auth_failed"},
		{ErrProviderBadRequest, "bad_request"},
		{ErrUserCancelled, "user_cancelled"},
		{errors.New("something else"), "unknown"},
	}
	for _, tt := range tests {
		got := LLMErrorCode(tt.err)
		if got != tt.want {
			t.Errorf("LLMErrorCode(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestLLMHTTPStatus(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{ErrProviderRateLimited, 429},
		{ErrProviderUnavailable, 503},
		{ErrProviderTimeout, 504},
		{ErrProviderAuthFailed, 401},
		{ErrProviderBadRequest, 400},
		{errors.New("other"), 0},
	}
	for _, tt := range tests {
		got := LLMHTTPStatus(tt.err)
		if got != tt.want {
			t.Errorf("LLMHTTPStatus(%v) = %d, want %d", tt.err, got, tt.want)
		}
	}
}
