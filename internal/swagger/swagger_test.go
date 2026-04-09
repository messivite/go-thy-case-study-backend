package swagger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesUI(t *testing.T) {
	h := Handler("/test-docs")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /: status %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("GET /: content-type %q", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "swagger-ui") {
		t.Fatal("GET /: missing swagger-ui reference")
	}
	if !strings.Contains(body, "/test-docs/openapi.json") {
		t.Fatalf("GET /: spec URL not embedded, body=%s", body)
	}
}

func TestHandlerServesYAML(t *testing.T) {
	h := Handler("/test-docs")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /openapi.yaml: status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Fatal("GET /openapi.yaml: missing openapi key")
	}
}

func TestHandlerServesJSON(t *testing.T) {
	h := Handler("/test-docs")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /openapi.json: status %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("GET /openapi.json: content-type %q", ct)
	}

	var spec map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &spec); err != nil {
		t.Fatalf("GET /openapi.json: invalid JSON: %v", err)
	}
	if _, ok := spec["openapi"]; !ok {
		t.Fatal("GET /openapi.json: missing openapi key")
	}
	if _, ok := spec["paths"]; !ok {
		t.Fatal("GET /openapi.json: missing paths key")
	}
}
