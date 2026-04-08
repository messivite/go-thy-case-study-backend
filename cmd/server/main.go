package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/messivite/go-thy-case-study-backend/internal/app"
	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/config"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func main() {
	port := envOrDefault("PORT", envOrDefault("APP_PORT", "8081"))
	roleClaimKey := envOrDefault("SUPABASE_ROLE_CLAIM_KEY", "role")

	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")
	validationMode := envOrDefault("SUPABASE_JWT_VALIDATION_MODE", "auto")

	authService := auth.NewSupabaseAuthAdapter(jwtSecret, supabaseURL, validationMode, roleClaimKey)

	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	var repository domain.Repository
	persistence := envOrDefault("CHAT_PERSISTENCE", "supabase")

	switch persistence {
	case "supabase":
		if supabaseURL == "" || serviceRoleKey == "" {
			log.Println("WARN: SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY empty, falling back to memory persistence")
			repository = repo.NewMemoryRepository()
		} else {
			repository = repo.NewSupabaseRepository(supabaseURL, serviceRoleKey)
			log.Println("chat persistence: supabase (postgres)")
		}
	default:
		repository = repo.NewMemoryRepository()
		log.Println("chat persistence: memory (in-process)")
	}

	registry := buildRegistry()

	uc := usecase.NewUseCase(repository, registry)
	chatHandler := chat.NewHandler(uc)

	server := app.NewServer(authService, chatHandler)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func buildRegistry() *provider.Registry {
	cfgPath := envOrDefault("PROVIDERS_CONFIG", config.DefaultConfigPath)
	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		log.Printf("WARN: providers.yaml not found (%v), using env-based fallback", err)
		return buildRegistryFromEnv()
	}

	if warnings := cfg.Validate(); len(warnings) > 0 {
		for _, w := range warnings {
			log.Printf("WARN: %s", w)
		}
	}

	registry := provider.NewRegistry(cfg.Default)
	for _, entry := range cfg.Providers {
		apiKey := os.Getenv(entry.EnvKey)
		if apiKey == "" {
			log.Printf("WARN: %s env key (%s) empty, provider disabled", entry.Name, entry.EnvKey)
			continue
		}
		p := createProvider(entry.Name, apiKey, entry.Model)
		if p == nil {
			log.Printf("WARN: unknown provider type %q, skipping", entry.Name)
			continue
		}
		registry.Register(p, provider.ProviderMeta{
			Name:           entry.Name,
			DefaultModel:   entry.Model,
			RequiredEnvKey: entry.EnvKey,
			SupportsStream: true,
		})
	}
	return registry
}

func buildRegistryFromEnv() *provider.Registry {
	registry := provider.NewRegistry("openai")
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		registry.Register(provider.NewOpenAIProvider(key, "gpt-4o"), provider.ProviderMeta{
			Name: "openai", DefaultModel: "gpt-4o", RequiredEnvKey: "OPENAI_API_KEY", SupportsStream: true,
		})
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		registry.Register(provider.NewGeminiProvider(key, "gemini-2.0-flash"), provider.ProviderMeta{
			Name: "gemini", DefaultModel: "gemini-2.0-flash", RequiredEnvKey: "GEMINI_API_KEY", SupportsStream: true,
		})
	}
	return registry
}

func createProvider(name, apiKey, model string) domain.LLMProvider {
	switch name {
	case "openai":
		return provider.NewOpenAIProvider(apiKey, model)
	case "gemini":
		return provider.NewGeminiProvider(apiKey, model)
	default:
		return nil
	}
}

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
