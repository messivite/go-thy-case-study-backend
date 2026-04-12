package app

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/landing"
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "ngrok-skip-browser-warning"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	health := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	r.Get("/", landing.Handler())
	r.Get("/health", health)

	if path := normalizeDocsPath(cfg.DocsPath); path != "" {
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, path+"/", http.StatusTemporaryRedirect)
		})
		r.Mount(path+"/", http.StripPrefix(path, swagger.Handler(path)))
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", health)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(authService))
			r.Get("/me", chatHandler.Me)
			r.Patch("/me", chatHandler.PatchMe)
			r.Get("/me/usage", chatHandler.MeUsage)
			r.Get("/chats/search", chatHandler.SearchChats)
			r.Get("/providers", chatHandler.ListProviders)
			r.Get("/models", chatHandler.ListModels)

			r.Post("/chats", chatHandler.CreateSession)
			r.Get("/chats", chatHandler.ListSessions)
			r.Get("/chats/{chatID}", chatHandler.GetChat)
			r.Delete("/chats/{chatID}", chatHandler.DeleteSession)
			r.Get("/chats/{chatID}/messages", chatHandler.ListMessages)
			r.Delete("/chats/{chatID}/messages/{messageID}", chatHandler.DeleteMessage)
			r.Post("/chats/{chatID}/messages/{messageID}/like", chatHandler.PostMessageLike)
			r.Post("/chats/{chatID}/likes/sync", chatHandler.PostSyncMessageLikes)
			r.Post("/chats/{chatID}/messages", chatHandler.PostMessage)
			r.Post("/chats/{chatID}/sync", chatHandler.SyncMessages)
			r.Post("/chats/{chatID}/stream", chatHandler.StreamMessage)
		})
	})

	return &Server{router: r}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func normalizeDocsPath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	if path == "" || path == "/" {
		return ""
	}
	return path
}
