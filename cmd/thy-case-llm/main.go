package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/thy-case-study-backend/internal/config"
)

const version = "0.1.0"

func main() {
	loadDotEnv()

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "provider":
		if len(args) < 2 {
			printProviderUsage()
			os.Exit(1)
		}
		handleProvider(args[1], args[2:])
	case "doctor":
		cmdDoctor()
	case "version":
		fmt.Printf("thy-case-llm v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func handleProvider(subcmd string, args []string) {
	switch subcmd {
	case "add":
		cmdProviderAdd(args)
	case "list":
		cmdProviderList()
	case "remove":
		cmdProviderRemove(args)
	case "set-default":
		cmdProviderSetDefault(args)
	case "validate":
		cmdProviderValidate()
	case "doctor":
		cmdDoctor()
	default:
		fmt.Fprintf(os.Stderr, "unknown provider subcommand: %s\n", subcmd)
		printProviderUsage()
		os.Exit(1)
	}
}

func cmdProviderAdd(args []string) {
	cfgPath := configPath()
	cfg := loadOrCreateConfig(cfgPath)

	var name, model, envKey string

	// Parse flags if provided
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "--model":
			if i+1 < len(args) {
				model = args[i+1]
				i++
			}
		case "--env-key":
			if i+1 < len(args) {
				envKey = args[i+1]
				i++
			}
		}
	}

	reader := bufio.NewReader(os.Stdin)

	if name == "" {
		fmt.Println("Desteklenen provider'lar: openai, gemini")
		fmt.Print("Provider adı: ")
		name = readLine(reader)
	}

	if model == "" {
		defaultModel := suggestModel(name)
		fmt.Printf("Model [%s]: ", defaultModel)
		model = readLine(reader)
		if model == "" {
			model = defaultModel
		}
	}

	if envKey == "" {
		defaultEnvKey := suggestEnvKey(name)
		fmt.Printf("Env key (API anahtarı için) [%s]: ", defaultEnvKey)
		envKey = readLine(reader)
		if envKey == "" {
			envKey = defaultEnvKey
		}
	}

	entry := config.ProviderEntry{
		Name:   name,
		Model:  model,
		EnvKey: envKey,
	}

	if err := cfg.AddProvider(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveProvidersConfig(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config kaydetme hatası: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Provider %q eklendi (model: %s, env: %s)\n", name, model, envKey)
	if cfg.Default == name {
		fmt.Printf("  → Varsayılan provider olarak ayarlandı\n")
	}
}

func cmdProviderList() {
	cfgPath := configPath()
	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Providers) == 0 {
		fmt.Println("Kayıtlı provider yok. `thy-case-llm provider add` ile ekleyin.")
		return
	}

	fmt.Printf("Varsayılan: %s\n\n", cfg.Default)
	fmt.Printf("%-15s %-20s %-25s %s\n", "ADI", "MODEL", "ENV KEY", "DURUM")
	fmt.Println(strings.Repeat("─", 70))

	for _, p := range cfg.Providers {
		status := "✗ env boş"
		if os.Getenv(p.EnvKey) != "" {
			status = "✓ aktif"
		}
		marker := "  "
		if p.Name == cfg.Default {
			marker = "→ "
		}
		fmt.Printf("%s%-13s %-20s %-25s %s\n", marker, p.Name, p.Model, p.EnvKey, status)
	}
}

func cmdProviderRemove(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Kullanım: thy-case-llm provider remove <name>")
		os.Exit(1)
	}
	name := args[0]
	cfgPath := configPath()
	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.RemoveProvider(name); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveProvidersConfig(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config kaydetme hatası: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Provider %q kaldırıldı\n", name)
}

func cmdProviderSetDefault(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Kullanım: thy-case-llm provider set-default <name>")
		os.Exit(1)
	}
	name := args[0]
	cfgPath := configPath()
	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.SetDefault(name); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveProvidersConfig(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config kaydetme hatası: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Varsayılan provider: %s\n", name)
}

func cmdProviderValidate() {
	cfgPath := configPath()
	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	warnings := cfg.Validate()
	if len(warnings) == 0 {
		fmt.Println("✓ Tüm provider'lar geçerli ve env anahtarları mevcut.")
		return
	}

	fmt.Println("Doğrulama sonuçları:")
	for _, w := range warnings {
		fmt.Printf("  ⚠ %s\n", w)
	}
	os.Exit(1)
}

func cmdDoctor() {
	cfgPath := configPath()
	failed := false

	fmt.Println("thy-case-llm doctor")
	fmt.Println(strings.Repeat("-", 48))

	if _, err := os.Stat(cfgPath); err != nil {
		fmt.Printf("✗ providers config bulunamadı: %s (%v)\n", cfgPath, err)
		failed = true
	} else {
		fmt.Printf("✓ providers config bulundu: %s\n", cfgPath)
	}

	cfg, err := config.LoadProvidersConfig(cfgPath)
	if err != nil {
		fmt.Printf("✗ providers config okunamadı: %v\n", err)
		os.Exit(1)
	}

	if cfg.Default == "" {
		fmt.Println("✗ default provider tanımlı değil")
		failed = true
	} else {
		fmt.Printf("✓ default provider: %s\n", cfg.Default)
	}

	if len(cfg.Providers) == 0 {
		fmt.Println("✗ providers listesi boş")
		failed = true
	} else {
		fmt.Printf("✓ kayıtlı provider sayısı: %d\n", len(cfg.Providers))
	}

	for _, p := range cfg.Providers {
		label := fmt.Sprintf("provider=%s model=%s", p.Name, p.Model)
		if p.EnvKey == "" {
			fmt.Printf("✗ %s -> env_key eksik\n", label)
			failed = true
			continue
		}
		if os.Getenv(p.EnvKey) == "" {
			fmt.Printf("✗ %s -> env boş (%s)\n", label, p.EnvKey)
			failed = true
			continue
		}
		fmt.Printf("✓ %s -> env hazır (%s)\n", label, p.EnvKey)
	}

	if _, err := os.Stat("api.yaml"); err != nil {
		fmt.Printf("⚠ api.yaml bulunamadı (%v)\n", err)
	} else {
		fmt.Println("✓ api.yaml bulundu")
	}

	if _, err := os.Stat(".env"); err != nil {
		fmt.Printf("⚠ .env bulunamadı (%v)\n", err)
	} else {
		fmt.Println("✓ .env bulundu")
	}

	if failed {
		fmt.Println("\nSonuç: bazı kontroller başarısız.")
		os.Exit(1)
	}
	fmt.Println("\nSonuç: tüm kritik kontroller başarılı.")
}

func loadOrCreateConfig(path string) *config.ProvidersConfig {
	cfg, err := config.LoadProvidersConfig(path)
	if err != nil {
		return &config.ProvidersConfig{}
	}
	return cfg
}

func configPath() string {
	if p := os.Getenv("PROVIDERS_CONFIG"); p != "" {
		return p
	}
	return config.DefaultConfigPath
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func suggestModel(providerName string) string {
	switch providerName {
	case "openai":
		return "gpt-4o"
	case "gemini":
		return "gemini-2.0-flash"
	default:
		return ""
	}
}

func suggestEnvKey(providerName string) string {
	switch providerName {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	default:
		return strings.ToUpper(providerName) + "_API_KEY"
	}
}

func printUsage() {
	fmt.Println(`thy-case-llm — LLM Provider Yönetim Aracı

Kullanım:
  thy-case-llm <komut> [argümanlar]

Komutlar:
  provider add                  Yeni bir LLM provider ekle
  provider list                 Kayıtlı provider'ları listele
  provider remove <name>        Provider'ı kaldır
  provider set-default <name>   Varsayılan provider'ı değiştir
  provider validate             Provider yapılandırmasını doğrula
  doctor                        Hızlı sistem sağlık kontrolü
  version                       Sürüm bilgisi
  help                          Bu yardım mesajı`)
}

func printProviderUsage() {
	fmt.Println(`Kullanım:
  thy-case-llm provider <alt-komut> [argümanlar]

Alt komutlar:
  add                           Yeni provider ekle (interaktif veya flag ile)
    --name <name>               Provider adı (openai, gemini, ...)
    --model <model>             Model adı (gpt-4o, gemini-2.0-flash, ...)
    --env-key <key>             API key ortam değişkeni adı
  list                          Kayıtlı provider'ları listele
  remove <name>                 Provider'ı kaldır
  set-default <name>            Varsayılan provider'ı değiştir
  validate                      Tüm provider'ları doğrula
  doctor                        Provider + env + config kontrolü`)
}

func loadDotEnv() {
	envPath := os.Getenv("ENV_FILE")
	if envPath == "" {
		envPath = ".env"
	}

	file, err := os.Open(filepath.Clean(envPath))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}

		// Keep explicitly exported shell values as source of truth.
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
