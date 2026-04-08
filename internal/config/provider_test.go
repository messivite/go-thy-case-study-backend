package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProvidersConfigAddRemove(t *testing.T) {
	cfg := &ProvidersConfig{}

	if err := cfg.AddProvider(ProviderEntry{Name: "openai", Model: "gpt-4o", EnvKey: "OPENAI_API_KEY"}); err != nil {
		t.Fatal(err)
	}
	if cfg.Default != "openai" {
		t.Errorf("expected default 'openai', got %q", cfg.Default)
	}

	if err := cfg.AddProvider(ProviderEntry{Name: "gemini", Model: "gemini-2.0-flash", EnvKey: "GEMINI_API_KEY"}); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(cfg.Providers))
	}

	if err := cfg.AddProvider(ProviderEntry{Name: "openai"}); err == nil {
		t.Error("expected error for duplicate provider")
	}

	if err := cfg.RemoveProvider("openai"); err != nil {
		t.Fatal(err)
	}
	if cfg.Default != "gemini" {
		t.Errorf("expected default to shift to 'gemini', got %q", cfg.Default)
	}

	if err := cfg.RemoveProvider("nonexistent"); err == nil {
		t.Error("expected error removing nonexistent provider")
	}
}

func TestProvidersConfigSetDefault(t *testing.T) {
	cfg := &ProvidersConfig{}
	cfg.AddProvider(ProviderEntry{Name: "a"})
	cfg.AddProvider(ProviderEntry{Name: "b"})

	if err := cfg.SetDefault("b"); err != nil {
		t.Fatal(err)
	}
	if cfg.Default != "b" {
		t.Errorf("expected default 'b', got %q", cfg.Default)
	}

	if err := cfg.SetDefault("nonexistent"); err == nil {
		t.Error("expected error for nonexistent default")
	}
}

func TestProvidersConfigValidate(t *testing.T) {
	cfg := &ProvidersConfig{
		Default: "openai",
		Providers: []ProviderEntry{
			{Name: "openai", Model: "gpt-4o", EnvKey: ""},
		},
	}

	warnings := cfg.Validate()
	if len(warnings) == 0 {
		t.Error("expected validation warnings for empty env_key")
	}
}

func TestProvidersConfigSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_providers.yaml")

	cfg := &ProvidersConfig{Default: "openai"}
	cfg.AddProvider(ProviderEntry{Name: "openai", Model: "gpt-4o", EnvKey: "OPENAI_API_KEY"})

	if err := SaveProvidersConfig(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadProvidersConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Default != "openai" {
		t.Errorf("expected default 'openai', got %q", loaded.Default)
	}
	if len(loaded.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(loaded.Providers))
	}
}

func TestProvidersConfigValidateEnvSet(t *testing.T) {
	t.Setenv("TEST_KEY_VALIDATE", "secretvalue")

	cfg := &ProvidersConfig{
		Default: "test",
		Providers: []ProviderEntry{
			{Name: "test", Model: "m", EnvKey: "TEST_KEY_VALIDATE"},
		},
	}

	warnings := cfg.Validate()
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
	_ = os.Getenv("TEST_KEY_VALIDATE")
}
