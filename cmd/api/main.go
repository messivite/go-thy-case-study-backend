package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/messivite/go-thy-case-study-backend/internal/app"
	usecase "github.com/messivite/go-thy-case-study-backend/internal/application/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	"github.com/messivite/go-thy-case-study-backend/internal/cache"
	"github.com/messivite/go-thy-case-study-backend/internal/catalog"
	"github.com/messivite/go-thy-case-study-backend/internal/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/config"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/dotenv"
	"github.com/messivite/go-thy-case-study-backend/internal/observability"
	"github.com/messivite/go-thy-case-study-backend/internal/provider"
	"github.com/messivite/go-thy-case-study-backend/internal/repo"
)

func main() {
	dotenv.LoadLocalEnv()

	if logPath := os.Getenv("OBSERVABILITY_LOG_FILE"); logPath != "" {
		if err := observability.EnableFileLog(logPath); err != nil {
			log.Fatalf("OBSERVABILITY_LOG_FILE: %v", err)
		}
		defer observability.CloseFileLog()
		log.Printf("observability: JSONL logs -> %s", logPath)
	}

	shutdownTracing, err := observability.InitTracing(context.Background())
	if err != nil {
		log.Fatalf("OpenTelemetry: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(ctx); err != nil {
			log.Printf("OpenTelemetry shutdown: %v", err)
		}
	}()

	port := envOrDefault("PORT", envOrDefault("APP_PORT", "8081"))
	roleClaimKey := envOrDefault("SUPABASE_ROLE_CLAIM_KEY", "role")

	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	supabaseURL := os.Getenv("SUPABASE_URL")
	validationMode := envOrDefault("SUPABASE_JWT_VALIDATION_MODE", "auto")

	authService := auth.NewSupabaseAuthAdapter(jwtSecret, supabaseURL, validationMode, roleClaimKey)

	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	var repository domain.Repository
	var modelCatalog domain.SupportedModelsCatalog
	persistence := strings.TrimSpace(envOrDefault("CHAT_PERSISTENCE", "supabase"))

	var quotaRepo domain.QuotaRepository

	switch persistence {
	case "supabase":
		if supabaseURL == "" || serviceRoleKey == "" {
			log.Println("WARN: SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY empty, falling back to memory persistence")
			mem := repo.NewMemoryRepository()
			repository = mem
			modelCatalog = mem
			quotaRepo = repo.NewMemoryQuotaRepository()
		} else {
			supabaseRepo := repo.NewSupabaseRepository(supabaseURL, serviceRoleKey)
			repository = supabaseRepo
			modelCatalog = supabaseRepo
			quotaRepo = supabaseRepo
			log.Println("chat persistence: supabase (postgres)")
		}
	default:
		mem := repo.NewMemoryRepository()
		repository = mem
		modelCatalog = mem
		quotaRepo = repo.NewMemoryQuotaRepository()
		log.Printf("chat persistence: memory (in-process) [CHAT_PERSISTENCE=%q]", persistence)
	}

	registry := buildRegistry()
	syncCtx, syncCancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := modelCatalog.SyncSupportedModels(syncCtx, catalog.SupportedModelsFromRegistry(registry)); err != nil {
		log.Printf("WARN: supported models sync: %v", err)
	}
	syncCancel()

	uc := usecase.NewUseCase(repository, quotaRepo, registry, modelCatalog)
	cacheStore, ttlList, ttlMsgs := cache.FromEnv()
	if cacheStore != nil && (ttlList > 0 || ttlMsgs > 0) {
		log.Printf("response cache: enabled (list TTL=%s, messages TTL=%s)", ttlList, ttlMsgs)
	}
	chatHandler := chat.NewHandler(uc, chat.WithResponseCache(cacheStore, ttlList, ttlMsgs))

	docsPath := os.Getenv("SWAGGER_PUBLIC_PATH")
	if docsPath == "" {
		docsPath = "/docs-a7b3c9e2f1d4"
	}
	serverCfg := app.ServerConfig{DocsPath: docsPath}
	log.Printf("swagger docs: http://localhost:%s%s", port, docsPath)

	server := app.NewServer(authService, chatHandler, serverCfg)
	handler := observability.HTTPHandler("thy-api", server.Handler())

	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
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
		registry.Register(provider.NewGeminiProvider(key, "gemini-2.5-flash"), provider.ProviderMeta{
			Name: "gemini", DefaultModel: "gemini-2.5-flash", RequiredEnvKey: "GEMINI_API_KEY", SupportsStream: true,
		})
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		registry.Register(provider.NewAnthropicProvider(key, "claude-sonnet-4-20250514"), provider.ProviderMeta{
			Name: "anthropic", DefaultModel: "claude-sonnet-4-20250514", RequiredEnvKey: "ANTHROPIC_API_KEY", SupportsStream: true,
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
	case "anthropic":
		return provider.NewAnthropicProvider(apiKey, model)
	case "claude":
		return provider.NewAnthropicProviderNamed(apiKey, model, "claude")
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
