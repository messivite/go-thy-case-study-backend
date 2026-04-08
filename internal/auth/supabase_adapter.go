package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	gosupabaseauth "github.com/messivite/gosupabase/auth"
	gosupabasemw "github.com/messivite/gosupabase/middleware"
)

type SupabaseAuthAdapter struct {
	jwtSecret      string
	supabaseURL    string
	validationMode string
	roleClaimKey   string
}

func NewSupabaseAuthAdapter(jwtSecret, supabaseURL, validationMode, roleClaimKey string) *SupabaseAuthAdapter {
	return &SupabaseAuthAdapter{
		jwtSecret:      jwtSecret,
		supabaseURL:    supabaseURL,
		validationMode: validationMode,
		roleClaimKey:   roleClaimKey,
	}
}

func (a *SupabaseAuthAdapter) AuthenticateRequest(r *http.Request) (*AuthenticatedUser, error) {
	middleware := gosupabasemw.SupabaseAuth(a.jwtSecret, a.supabaseURL, a.validationMode)
	var validatedReq *http.Request

	recorder := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		validatedReq = req
	})).ServeHTTP(recorder, r)

	if validatedReq == nil {
		return nil, errors.New("unauthorized")
	}

	claims := gosupabaseauth.GetClaims(validatedReq.Context())
	if claims == nil {
		return nil, errors.New("missing claims")
	}

	if claims.Subject == "" {
		return nil, errors.New("missing user identifier in token")
	}

	return &AuthenticatedUser{
		UserID: claims.Subject,
		Email:  claims.Email,
		Role:   a.extractRole(validatedReq, claims),
	}, nil
}

func (a *SupabaseAuthAdapter) extractRole(r *http.Request, claims *gosupabaseauth.Claims) string {
	if a.roleClaimKey == "" || strings.EqualFold(a.roleClaimKey, "role") {
		return claims.Role
	}
	if strings.EqualFold(a.roleClaimKey, "roles") {
		if len(claims.Roles) > 0 {
			return claims.Roles[0]
		}
		return ""
	}

	role, err := extractClaimFromToken(r, a.roleClaimKey)
	if err != nil {
		return ""
	}
	return role
}

func extractClaimFromToken(r *http.Request, claimKey string) (string, error) {
	token, err := ExtractBearerToken(r)
	if err != nil {
		return "", err
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token format")
	}

	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return "", err
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", err
	}

	value, ok := raw[claimKey]
	if !ok {
		return "", nil
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []any:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str, nil
			}
		}
	}

	return fmt.Sprintf("%v", value), nil
}

func base64URLDecode(s string) ([]byte, error) {
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}
