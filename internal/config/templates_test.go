package config

import (
	"testing"
)

func TestBuiltinTemplatesExist(t *testing.T) {
	expected := []string{"openai", "gemini", "anthropic", "claude"}
	for _, name := range expected {
		if _, ok := BuiltinTemplates[name]; !ok {
			t.Errorf("expected built-in template %q to exist", name)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	tpl, ok := GetTemplate("openai")
	if !ok {
		t.Fatal("expected to find openai template")
	}
	if tpl.Name != "openai" {
		t.Errorf("expected name 'openai', got %q", tpl.Name)
	}
	if tpl.DefaultModel == "" {
		t.Error("expected non-empty default model")
	}
	if len(tpl.Models) == 0 {
		t.Error("expected non-empty models list")
	}
}

func TestGetTemplateNotFound(t *testing.T) {
	_, ok := GetTemplate("nonexistent")
	if ok {
		t.Error("expected template not found")
	}
}

func TestIsKnownTemplate(t *testing.T) {
	if !IsKnownTemplate("gemini") {
		t.Error("expected gemini to be known")
	}
	if !IsKnownTemplate("claude") {
		t.Error("expected claude template alias to be known")
	}
}

func TestListTemplateNames(t *testing.T) {
	names := ListTemplateNames()
	if len(names) < 3 {
		t.Errorf("expected at least 3 templates, got %d", len(names))
	}
}

func TestTemplateFieldsNotEmpty(t *testing.T) {
	for name, tpl := range BuiltinTemplates {
		if tpl.DisplayName == "" {
			t.Errorf("%s: display name empty", name)
		}
		if tpl.EnvKey == "" {
			t.Errorf("%s: env key empty", name)
		}
		if tpl.BaseURL == "" {
			t.Errorf("%s: base URL empty", name)
		}
		if tpl.AuthType == "" {
			t.Errorf("%s: auth type empty", name)
		}
		if tpl.Description == "" {
			t.Errorf("%s: description empty", name)
		}
	}
}
