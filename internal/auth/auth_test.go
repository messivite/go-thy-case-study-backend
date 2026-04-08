package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubAuthService struct {
	user *AuthenticatedUser
	err  error
}

func (s stubAuthService) AuthenticateRequest(r *http.Request) (*AuthenticatedUser, error) {
	return s.user, s.err
}

func TestExtractBearerToken(t *testing.T) {
	t.Run("valid bearer token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer test-token")

		token, err := ExtractBearerToken(r)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token != "test-token" {
			t.Fatalf("expected test-token, got %s", token)
		}
	})

	t.Run("missing header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if _, err := ExtractBearerToken(r); err == nil {
			t.Fatal("expected error for missing authorization header")
		}
	})

	t.Run("invalid scheme", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Basic abc")
		if _, err := ExtractBearerToken(r); err == nil {
			t.Fatal("expected error for invalid authorization header")
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := AuthenticatedUserFromContext(r.Context())
		if !ok || user == nil {
			t.Fatal("expected user in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("authorized request passes through", func(t *testing.T) {
		service := stubAuthService{user: &AuthenticatedUser{UserID: "u1"}}
		handler := AuthMiddleware(service)(next)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("unauthorized request blocked", func(t *testing.T) {
		service := stubAuthService{err: context.DeadlineExceeded}
		handler := AuthMiddleware(service)(next)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
