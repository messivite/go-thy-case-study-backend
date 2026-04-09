package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/messivite/go-thy-case-study-backend/internal/auth"
)

type noopAuth struct{}

func (noopAuth) AuthenticateRequest(r *http.Request) (*auth.AuthenticatedUser, error) {
	return &auth.AuthenticatedUser{UserID: "test"}, nil
}

func TestHealthRoutesNoAuth(t *testing.T) {
	s := NewServer(noopAuth{}, nil, ServerConfig{})

	for _, path := range []string{"/health", "/api/health"} {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			s.Handler().ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("GET %s: status %d, body %q", path, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestDocsRouteNoAuth(t *testing.T) {
	cfg := ServerConfig{DocsPath: "/test-docs-secret"}
	s := NewServer(noopAuth{}, nil, cfg)

	paths := []string{
		"/test-docs-secret/",
		"/test-docs-secret/openapi.json",
		"/test-docs-secret/openapi.yaml",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			s.Handler().ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("GET %s: status %d, body %q", path, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestDocsDisabledWhenEmpty(t *testing.T) {
	s := NewServer(noopAuth{}, nil, ServerConfig{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs-a7b3c9e2f1d4", nil)
	s.Handler().ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatal("docs should not be mounted when DocsPath is empty")
	}
}
