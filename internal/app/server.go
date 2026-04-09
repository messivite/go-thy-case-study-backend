package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/chat"
)

type Server struct {
	router http.Handler
}

func NewServer(authService auth.AuthService, chatHandler *chat.Handler) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	health := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	// Kök: Dockerfile / Railway / curl …:8081/health
	r.Get("/health", health)

	r.Route("/api", func(r chi.Router) {
		// basePath /api + /health (Postman vb.) — JWT yok; auth sadece alt grupta
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
