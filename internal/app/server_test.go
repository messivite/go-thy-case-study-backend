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
	s := NewServer(noopAuth{}, nil)

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
