package auth

import (
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

	user := &AuthenticatedUser{
		UserID: claims.Subject,
		Email:  claims.Email,
		Role:   a.extractRole(validatedReq, claims),
	}
	raw, err := ParseAccessTokenPayload(r)
	if err == nil {
		user.JWTClaims = raw
		applyJWTExtras(user, claims, raw)
	} else {
		applyJWTExtras(user, claims, nil)
	}
	return user, nil
}

func applyJWTExtras(u *AuthenticatedUser, claims *gosupabaseauth.Claims, raw map[string]any) {
	if claims != nil {
		u.Roles = claims.EffectiveRoles()
		u.Audience = claims.Audience
		u.ExpiresAt = claims.ExpiresAt
	}
	if raw == nil {
		return
	}
	if v, ok := jwtNumericInt64(raw, "iat"); ok {
		u.IssuedAt = v
	}
	if v, ok := jwtNumericInt64(raw, "exp"); ok && u.ExpiresAt == 0 {
		u.ExpiresAt = v
	}
	if s, ok := raw["iss"].(string); ok {
		u.Issuer = s
	}
	if s, ok := raw["phone"].(string); ok {
		u.Phone = s
	}
	if s, ok := raw["session_id"].(string); ok {
		u.SessionID = s
	}
	if m, ok := metadataMap(raw["app_metadata"]); ok {
		u.AppMetadata = m
	}
	if m, ok := metadataMap(raw["user_metadata"]); ok {
		u.UserMetadata = m
	}
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
	raw, err := ParseAccessTokenPayload(r)
	if err != nil {
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
