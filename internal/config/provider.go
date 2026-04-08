package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ProviderEntry struct {
	Name   string `yaml:"name"`
	Model  string `yaml:"model"`
	EnvKey string `yaml:"env_key"`
}

type ProvidersConfig struct {
	Default   string          `yaml:"default"`
	Providers []ProviderEntry `yaml:"providers"`
}

const DefaultConfigPath = "providers.yaml"

func LoadProvidersConfig(path string) (*ProvidersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read providers config: %w", err)
	}
	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse providers config: %w", err)
	}
	return &cfg, nil
}

func SaveProvidersConfig(path string, cfg *ProvidersConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal providers config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (c *ProvidersConfig) FindProvider(name string) (*ProviderEntry, int) {
	for i := range c.Providers {
		if c.Providers[i].Name == name {
			return &c.Providers[i], i
		}
	}
	return nil, -1
}

func (c *ProvidersConfig) AddProvider(entry ProviderEntry) error {
	if existing, _ := c.FindProvider(entry.Name); existing != nil {
		return fmt.Errorf("provider %q already exists", entry.Name)
	}
	c.Providers = append(c.Providers, entry)
	if c.Default == "" {
		c.Default = entry.Name
	}
	return nil
}

func (c *ProvidersConfig) RemoveProvider(name string) error {
	_, idx := c.FindProvider(name)
	if idx < 0 {
		return fmt.Errorf("provider %q not found", name)
	}
	c.Providers = append(c.Providers[:idx], c.Providers[idx+1:]...)
	if c.Default == name {
		if len(c.Providers) > 0 {
			c.Default = c.Providers[0].Name
		} else {
			c.Default = ""
		}
	}
	return nil
}

func (c *ProvidersConfig) SetDefault(name string) error {
	if p, _ := c.FindProvider(name); p == nil {
		return fmt.Errorf("provider %q not found", name)
	}
	c.Default = name
	return nil
}

func (c *ProvidersConfig) Validate() []string {
	var warnings []string
	for _, p := range c.Providers {
		if p.EnvKey == "" {
			warnings = append(warnings, fmt.Sprintf("[%s] env_key is not set", p.Name))
			continue
		}
		if os.Getenv(p.EnvKey) == "" {
			warnings = append(warnings, fmt.Sprintf("[%s] env var %s is empty", p.Name, p.EnvKey))
		}
	}
	if c.Default != "" {
		if p, _ := c.FindProvider(c.Default); p == nil {
			warnings = append(warnings, fmt.Sprintf("default provider %q is not in provider list", c.Default))
		}
	}
	return warnings
}
