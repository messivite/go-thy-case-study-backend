package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/example/thy-case-study-backend/internal/app"
	"github.com/example/thy-case-study-backend/internal/auth"
	"github.com/example/thy-case-study-backend/internal/chat"
	"github.com/example/thy-case-study-backend/internal/provider"
	"github.com/example/thy-case-study-backend/internal/repo"
)

func main() {
	port := envOrDefault("PORT", envOrDefault("APP_PORT", "8081"))
	roleClaimKey := envOrDefault("SUPABASE_ROLE_CLAIM_KEY", "role")
	openAIKey := os.Getenv("OPENAI_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")
	validationMode := envOrDefault("SUPABASE_JWT_VALIDATION_MODE", "auto")

	authService := auth.NewSupabaseAuthAdapter(jwtSecret, supabaseURL, validationMode, roleClaimKey)

	repository := repo.NewMemoryRepository()

	providers := []provider.LLMProvider{
		provider.NewOpenAIProvider(openAIKey),
		provider.NewGeminiProvider(geminiKey),
	}
	providerFactory := provider.NewProviderFactory(providers)

	chatService := chat.NewChatService(repository, providerFactory)
	chatHandler := chat.NewHandler(chatService)

	server := app.NewServer(authService, chatHandler)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func envOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
