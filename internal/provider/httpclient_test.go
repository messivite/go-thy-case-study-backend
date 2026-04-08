package provider

import (
	"errors"
	"net/http"
	"testing"

	domain "github.com/example/thy-case-study-backend/internal/domain/chat"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		status   int
		wantErr  error
		provider string
	}{
		{401, domain.ErrProviderAuthFailed, "openai"},
		{403, domain.ErrProviderAuthFailed, "openai"},
		{429, domain.ErrProviderRateLimited, "gemini"},
		{400, domain.ErrProviderBadRequest, "openai"},
		{500, domain.ErrProviderUnavailable, "openai"},
		{502, domain.ErrProviderUnavailable, "openai"},
		{503, domain.ErrProviderUnavailable, "gemini"},
		{408, domain.ErrProviderTimeout, "openai"},
		{504, domain.ErrProviderUnavailable, "openai"},
	}

	for _, tt := range tests {
		err := MapHTTPError(tt.status, tt.provider)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("MapHTTPError(%d, %s) = %v, want %v", tt.status, tt.provider, err, tt.wantErr)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	if !IsRetryable(500) {
		t.Error("500 should be retryable")
	}
	if !IsRetryable(503) {
		t.Error("503 should be retryable")
	}
	if !IsRetryable(http.StatusTooManyRequests) {
		t.Error("429 should be retryable")
	}
	if IsRetryable(200) {
		t.Error("200 should not be retryable")
	}
	if IsRetryable(400) {
		t.Error("400 should not be retryable")
	}
}
