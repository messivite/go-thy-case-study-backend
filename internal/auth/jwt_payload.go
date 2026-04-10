package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// ParseAccessTokenPayload decodes the JWT payload (middle segment) into a map.
// Numbers appear as float64 (standard encoding/json behavior for map[string]any).
func ParseAccessTokenPayload(r *http.Request) (map[string]any, error) {
	token, err := ExtractBearerToken(r)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	payload, err := decodeBase64URL(parts[1])
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func decodeBase64URL(s string) ([]byte, error) {
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

func jwtNumericInt64(m map[string]any, key string) (int64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case json.Number:
		i, err := x.Int64()
		return i, err == nil
	}
	return 0, false
}

func metadataMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}
