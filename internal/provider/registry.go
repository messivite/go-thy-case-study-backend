package provider

import (
	"fmt"

	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
)

type ProviderMeta struct {
	Name           string `json:"name" yaml:"name"`
	DefaultModel   string `json:"defaultModel" yaml:"model"`
	RequiredEnvKey string `json:"requiredEnvKey" yaml:"env_key"`
	SupportsStream bool   `json:"supportsStream" yaml:"-"`
	Enabled        bool   `json:"enabled" yaml:"-"`
}

type Registry struct {
	providers   map[string]domain.LLMProvider
	meta        map[string]ProviderMeta
	defaultName string
}

func NewRegistry(defaultName string) *Registry {
	return &Registry{
		providers:   make(map[string]domain.LLMProvider),
		meta:        make(map[string]ProviderMeta),
		defaultName: defaultName,
	}
}

func (r *Registry) Register(p domain.LLMProvider, meta ProviderMeta) {
	meta.Enabled = true
	r.providers[p.Name()] = p
	r.meta[p.Name()] = meta
}

func (r *Registry) Get(name string) (domain.LLMProvider, error) {
	if name == "" {
		name = r.defaultName
	}
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domain.ErrUnsupportedProvider, name)
	}
	return p, nil
}

func (r *Registry) List() []ProviderMeta {
	out := make([]ProviderMeta, 0, len(r.meta))
	for _, m := range r.meta {
		out = append(out, m)
	}
	return out
}

func (r *Registry) ListNames() []string {
	out := make([]string, 0, len(r.providers))
	for name := range r.providers {
		out = append(out, name)
	}
	return out
}

func (r *Registry) Default() string {
	return r.defaultName
}

// Meta returns registered metadata for a provider. Empty name uses the registry default provider.
func (r *Registry) Meta(name string) (ProviderMeta, bool) {
	if name == "" {
		name = r.defaultName
	}
	m, ok := r.meta[name]
	return m, ok
}

func (r *Registry) SetDefault(name string) error {
	if _, ok := r.providers[name]; !ok {
		return fmt.Errorf("%w: %s", domain.ErrUnsupportedProvider, name)
	}
	r.defaultName = name
	return nil
}
