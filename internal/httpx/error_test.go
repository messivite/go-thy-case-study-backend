package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusBadRequest, "bad_request", "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var env ErrorEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}

	if env.Error.Code != "bad_request" {
		t.Errorf("expected code 'bad_request', got %q", env.Error.Code)
	}
	if env.Error.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got %q", env.Error.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "not found")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestInternal(t *testing.T) {
	w := httptest.NewRecorder()
	Internal(w)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestProviderErrors(t *testing.T) {
	tests := []struct {
		fn   func(http.ResponseWriter)
		code int
	}{
		{ProviderAuthFailed, http.StatusBadGateway},
		{ProviderTimeout, http.StatusGatewayTimeout},
		{ProviderRateLimited, http.StatusTooManyRequests},
		{ProviderUnavailable, http.StatusBadGateway},
		{QuotaDailyExceeded, http.StatusTooManyRequests},
		{QuotaWeeklyExceeded, http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		tt.fn(w)
		if w.Code != tt.code {
			t.Errorf("expected %d, got %d", tt.code, w.Code)
		}

		var env ErrorEnvelope
		if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
			t.Fatal(err)
		}
		if env.Error.Code == "" {
			t.Error("expected non-empty error code")
		}
	}
}
