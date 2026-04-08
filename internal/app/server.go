package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/example/thy-case-study-backend/internal/auth"
	"github.com/example/thy-case-study-backend/internal/chat"
)

type Server struct {
	router http.Handler
}

func NewServer(authService auth.AuthService, chatHandler *chat.Handler) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(authService))
			r.Get("/me", chatHandler.Me)
		r.Post("/sessions", chatHandler.CreateSession)
		r.Get("/sessions/{sessionID}/messages", chatHandler.ListMessages)
		r.Post("/sessions/{sessionID}/messages", chatHandler.PostMessage)
	})

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
