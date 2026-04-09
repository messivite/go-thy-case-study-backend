package app

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/swagger"
)

type ServerConfig struct {
	DocsPath string // e.g. "/docs-a7b3c9e2f1d4"; empty disables docs
}

type Server struct {
	router http.Handler
}

func NewServer(authService auth.AuthService, chatHandler *chat.Handler, cfg ServerConfig) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	health := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	r.Get("/health", health)

	if path := strings.TrimRight(cfg.DocsPath, "/"); path != "" {
		r.Mount(path, http.StripPrefix(path, swagger.Handler(path)))
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", health)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(authService))
			r.Get("/me", chatHandler.Me)
			r.Get("/providers", chatHandler.ListProviders)

			r.Post("/chats", chatHandler.CreateSession)
			r.Get("/chats", chatHandler.ListSessions)
			r.Get("/chats/{chatID}", chatHandler.GetChat)
			r.Post("/chats/{chatID}/messages", chatHandler.PostMessage)
			r.Post("/chats/{chatID}/stream", chatHandler.StreamMessage)
		})
	})

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
