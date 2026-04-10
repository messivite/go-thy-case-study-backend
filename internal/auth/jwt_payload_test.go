package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseAccessTokenPayload(t *testing.T) {
	payload := map[string]any{
		"sub":   "550e8400-e29b-41d4-a716-446655440000",
		"email": "a@b.co",
		"exp":   float64(1893456000),
		"iat":   float64(1893450000),
		"app_metadata": map[string]any{
			"roles": []any{"editor"},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding.EncodeToString(b)
	token := "e30." + enc + ".sig"
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	got, err := ParseAccessTokenPayload(r)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got["sub"] != payload["sub"] {
		t.Fatalf("sub: %v", got["sub"])
	}
	v, ok := jwtNumericInt64(got, "exp")
	if !ok || v != 1893456000 {
		t.Fatalf("exp: ok=%v v=%d", ok, v)
	}
}

func TestJWTNumericInt64(t *testing.T) {
	m := map[string]any{"x": float64(42)}
	v, ok := jwtNumericInt64(m, "x")
	if !ok || v != 42 {
		t.Fatalf("got %d ok=%v", v, ok)
	}
	if _, ok := jwtNumericInt64(m, "missing"); ok {
		t.Fatal("expected miss")
	}
}

func TestDecodeBase64URL(t *testing.T) {
	s := strings.TrimRight(base64.URLEncoding.EncodeToString([]byte("hi")), "=")
	b, err := decodeBase64URL(s)
	if err != nil || string(b) != "hi" {
		t.Fatalf("decode: %v %q", err, b)
	}
}
