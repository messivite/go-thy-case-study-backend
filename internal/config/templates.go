package config

type ProviderTemplate struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"displayName"`
	DefaultModel   string   `json:"defaultModel"`
	Models         []string `json:"models"`
	EnvKey         string   `json:"envKey"`
	SupportsStream bool     `json:"supportsStream"`
	BaseURL        string   `json:"baseURL"`
	AuthType       string   `json:"authType"`
	Description    string   `json:"description"`
}

var BuiltinTemplates = map[string]ProviderTemplate{
	"openai": {
		Name:           "openai",
		DisplayName:    "OpenAI",
		DefaultModel:   "gpt-4o",
		Models:         []string{"gpt-4o", "gpt-4o-mini", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "o4-mini"},
		EnvKey:         "OPENAI_API_KEY",
		SupportsStream: true,
		BaseURL:        "https://api.openai.com/v1",
		AuthType:       "bearer",
		Description:    "OpenAI ChatGPT modelleri (GPT-4o, GPT-4.1 serisi)",
	},
	"gemini": {
		Name:           "gemini",
		DisplayName:    "Google Gemini",
		DefaultModel:   "gemini-2.5-flash",
		Models:         []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.5-flash-lite"},
		EnvKey:         "GEMINI_API_KEY",
		SupportsStream: true,
		BaseURL:        "https://generativelanguage.googleapis.com/v1beta",
		AuthType:       "query-key",
		Description:    "Google Gemini modelleri (2.5 Flash / Pro / Flash-Lite; 2.0 yeni hesaplarda kapalı)",
	},
	"anthropic": {
		Name:           "anthropic",
		DisplayName:    "Anthropic Claude",
		DefaultModel:   "claude-sonnet-4-20250514",
		Models:         []string{"claude-sonnet-4-20250514", "claude-4-opus-20250514", "claude-3.5-haiku-20241022"},
		EnvKey:         "ANTHROPIC_API_KEY",
		SupportsStream: true,
		BaseURL:        "https://api.anthropic.com/v1",
		AuthType:       "x-api-key",
		Description:    "Anthropic Claude modelleri (Claude 4 Sonnet/Opus, 3.5 Haiku)",
	},
	// claude: providers.yaml'da name: claude; aynı API ve ANTHROPIC_API_KEY (createProvider "claude" dalı).
	"claude": {
		Name:           "claude",
		DisplayName:    "Claude (Anthropic API)",
		DefaultModel:   "claude-sonnet-4-20250514",
		Models:         []string{"claude-sonnet-4-20250514", "claude-4-opus-20250514", "claude-3.5-haiku-20241022"},
		EnvKey:         "ANTHROPIC_API_KEY",
		SupportsStream: true,
		BaseURL:        "https://api.anthropic.com/v1",
		AuthType:       "x-api-key",
		Description:    "Anthropic Messages API — Claude modelleri (providers.yaml kaydı anthropic yerine claude olur)",
	},
}

func ListTemplateNames() []string {
	names := make([]string, 0, len(BuiltinTemplates))
	for name := range BuiltinTemplates {
		names = append(names, name)
	}
	return names
}

func GetTemplate(name string) (ProviderTemplate, bool) {
	t, ok := BuiltinTemplates[name]
	return t, ok
}

func IsKnownTemplate(name string) bool {
	_, ok := BuiltinTemplates[name]
	return ok
}
