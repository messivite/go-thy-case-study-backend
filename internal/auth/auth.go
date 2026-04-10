package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/messivite/go-thy-case-study-backend/internal/httpx"
)

// AuthenticatedUser is filled by AuthService (e.g. Supabase JWT). Me handler
// maps this to a richer JSON shape; handlers should use UserID / Email / Role for RBAC.
type AuthenticatedUser struct {
	UserID       string
	Email        string
	Role         string
	Roles        []string
	Issuer       string
	Audience     string
	IssuedAt     int64
	ExpiresAt    int64
	Phone        string
	SessionID    string
	AppMetadata  map[string]any
	UserMetadata map[string]any
	// JWTClaims is the full decoded access-token payload (registered + custom claims).
	JWTClaims map[string]any
}

type AuthService interface {
	AuthenticateRequest(r *http.Request) (*AuthenticatedUser, error)
}

type contextKey string

const authContextKey contextKey = "authenticatedUser"

func ContextWithAuthenticatedUser(ctx context.Context, user *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authContextKey, user)
}

func AuthenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(authContextKey).(*AuthenticatedUser)
	return user, ok
}

func AuthMiddleware(authService AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authService.AuthenticateRequest(r)
			if err != nil {
				httpx.Unauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(ContextWithAuthenticatedUser(r.Context(), user)))
		})
	}
}

func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}

	return strings.TrimSpace(parts[1]), nil
}
