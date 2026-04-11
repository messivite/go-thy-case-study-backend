package chat

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func TestHandler_PatchMe_JSON(t *testing.T) {
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := auth.ContextWithAuthenticatedUser(req.Context(), &auth.AuthenticatedUser{UserID: "u-json"})
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	r.Patch("/api/me", h.PatchMe)

	body := `{"displayName":"Pat","locale":"en"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var out meResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Profile.DisplayName != "Pat" || out.Profile.Locale != "en" {
		t.Fatalf("profile %+v", out.Profile)
	}
}

func TestHandler_PatchMe_multipartAvatar(t *testing.T) {
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := auth.ContextWithAuthenticatedUser(req.Context(), &auth.AuthenticatedUser{UserID: "u-multi"})
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	r.Patch("/api/me", h.PatchMe)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("displayName", "Sam")
	part, err := w.CreateFormFile("avatar", "x.png")
	if err != nil {
		t.Fatal(err)
	}
	// minimal 1x1 PNG
	png1x1 := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if _, err := part.Write(png1x1); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/me", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var out meResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Profile.DisplayName != "Sam" || !strings.Contains(out.Profile.AvatarURL, "u-multi") {
		t.Fatalf("profile %+v", out.Profile)
	}
}

func TestHandler_PatchMe_emptyBody(t *testing.T) {
	mem := repo.NewMemoryRepository()
	reg := provider.NewRegistry("x")
	uc := usecase.NewUseCase(mem, handlerTestQuotaStub{}, reg, mem)
	h := NewHandler(uc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := auth.ContextWithAuthenticatedUser(req.Context(), &auth.AuthenticatedUser{UserID: "u-empty"})
			next.ServeHTTP(w, req.WithContext(c))
		})
	})
	r.Patch("/api/me", h.PatchMe)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", rr.Code, rr.Body.String())
	}
}
