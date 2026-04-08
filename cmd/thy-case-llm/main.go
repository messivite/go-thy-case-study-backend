package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/messivite/go-thy-case-study-backend/internal/config"
	"github.com/messivite/go-thy-case-study-backend/internal/deploy"
)

const version = "0.3.0"

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
	case "templates":
		if len(args) < 2 {
			printTemplatesUsage()
			os.Exit(1)
		}
		handleTemplates(args[1], args[2:])
	case "deploy":
		if len(args) < 2 {
			printDeployUsage()
			os.Exit(1)
		}
		handleDeploy(args[1], args[2:])
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

// ---------------------------------------------------------------------------
// provider subcommands
// ---------------------------------------------------------------------------

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
	case "templates":
		if len(args) < 1 {
			printTemplatesUsage()
			os.Exit(1)
		}
		handleTemplates(args[0], args[1:])
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

	var name, model, envKey, template string
	var setDefault bool

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
		case "--template":
			if i+1 < len(args) {
				template = args[i+1]
				i++
			}
		case "--set-default":
			setDefault = true
		}
	}

	reader := bufio.NewReader(os.Stdin)

	if template != "" {
		tpl, ok := config.GetTemplate(template)
		if !ok {
			fmt.Fprintf(os.Stderr, "Bilinmeyen template: %q\n", template)
			fmt.Fprintf(os.Stderr, "Mevcut template'ler: %s\n", strings.Join(sortedTemplateNames(), ", "))
			os.Exit(1)
		}
		if name == "" {
			name = tpl.Name
		}
		if model == "" {
			model = tpl.DefaultModel
		}
		if envKey == "" {
			envKey = tpl.EnvKey
		}
		fmt.Printf("Template: %s (%s)\n", tpl.DisplayName, tpl.Description)
	} else {
		if name == "" {
			fmt.Printf("Mevcut template'ler: %s\n", strings.Join(sortedTemplateNames(), ", "))
			fmt.Print("Provider adı (veya template adı): ")
			name = readLine(reader)
		}

		if tpl, ok := config.GetTemplate(name); ok {
			template = name
			if model == "" {
				model = tpl.DefaultModel
			}
			if envKey == "" {
				envKey = tpl.EnvKey
			}
			fmt.Printf("Template bulundu: %s\n", tpl.DisplayName)
		}
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
		fmt.Printf("Env key [%s]: ", defaultEnvKey)
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

	if setDefault {
		cfg.Default = name
	}

	if err := config.SaveProvidersConfig(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Config kaydetme hatası: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Provider %q eklendi (model: %s, env: %s)\n", name, model, envKey)
	if cfg.Default == name {
		fmt.Printf("  → Varsayılan provider olarak ayarlandı\n")
	}

	if !config.IsKnownTemplate(name) {
		fmt.Println()
		fmt.Println("  ⚠ Bu provider için hazır adapter bulunamadı.")
		fmt.Println("  Custom adapter eklemek için:")
		fmt.Println("    1) internal/provider/<name>.go dosyası oluşturun")
		fmt.Println("    2) domain.LLMProvider interface'ini implemente edin")
		fmt.Println("    3) cmd/api/main.go createProvider() switch'ine ekleyin")
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
	fmt.Printf("%-15s %-20s %-25s %-12s %s\n", "ADI", "MODEL", "ENV KEY", "TEMPLATE", "DURUM")
	fmt.Println(strings.Repeat("─", 85))

	for _, p := range cfg.Providers {
		status := "✗ env boş"
		if os.Getenv(p.EnvKey) != "" {
			status = "✓ aktif"
		}
		marker := "  "
		if p.Name == cfg.Default {
			marker = "→ "
		}
		tplTag := "custom"
		if config.IsKnownTemplate(p.Name) {
			tplTag = "built-in"
		}
		fmt.Printf("%s%-13s %-20s %-25s %-12s %s\n", marker, p.Name, p.Model, p.EnvKey, tplTag, status)
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

// ---------------------------------------------------------------------------
// templates subcommands
// ---------------------------------------------------------------------------

func handleTemplates(subcmd string, args []string) {
	switch subcmd {
	case "list":
		cmdTemplatesList()
	case "show":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Kullanım: thy-case-llm templates show <name>")
			os.Exit(1)
		}
		cmdTemplatesShow(args[0])
	default:
		fmt.Fprintf(os.Stderr, "unknown templates subcommand: %s\n", subcmd)
		printTemplatesUsage()
		os.Exit(1)
	}
}

func cmdTemplatesList() {
	names := sortedTemplateNames()

	fmt.Println("Mevcut Provider Template'leri")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("%-15s %-20s %-20s %s\n", "ADI", "VARSAYILAN MODEL", "STREAM", "AÇIKLAMA")
	fmt.Println(strings.Repeat("─", 70))

	for _, name := range names {
		tpl := config.BuiltinTemplates[name]
		stream := "✓"
		if !tpl.SupportsStream {
			stream = "✗"
		}
		fmt.Printf("%-15s %-20s %-20s %s\n", tpl.Name, tpl.DefaultModel, stream, tpl.Description)
	}

	fmt.Println()
	fmt.Println("Kullanım: thy-case-llm provider add --template <name>")
	fmt.Println("Detay:    thy-case-llm templates show <name>")
}

func cmdTemplatesShow(name string) {
	tpl, ok := config.GetTemplate(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Template bulunamadı: %q\n", name)
		fmt.Fprintf(os.Stderr, "Mevcut: %s\n", strings.Join(sortedTemplateNames(), ", "))
		os.Exit(1)
	}

	fmt.Printf("Template: %s\n", tpl.DisplayName)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Ad:            %s\n", tpl.Name)
	fmt.Printf("  Açıklama:      %s\n", tpl.Description)
	fmt.Printf("  Base URL:      %s\n", tpl.BaseURL)
	fmt.Printf("  Auth tipi:     %s\n", tpl.AuthType)
	fmt.Printf("  Env key:       %s\n", tpl.EnvKey)
	fmt.Printf("  Varsayılan:    %s\n", tpl.DefaultModel)
	fmt.Printf("  Stream:        %v\n", tpl.SupportsStream)
	fmt.Printf("  Modeller:      %s\n", strings.Join(tpl.Models, ", "))

	fmt.Println()
	fmt.Printf("Hızlı ekleme:\n")
	fmt.Printf("  thy-case-llm provider add --template %s\n", tpl.Name)
	fmt.Printf("  thy-case-llm provider add --template %s --model %s --set-default\n", tpl.Name, tpl.DefaultModel)
}

// ---------------------------------------------------------------------------
// deploy
// ---------------------------------------------------------------------------

func handleDeploy(subcmd string, args []string) {
	switch subcmd {
	case "list":
		if err := deploy.List(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	case "show":
		if len(args) < 1 {
			printDeployUsage()
			os.Exit(1)
		}
		if err := deploy.Show(os.Stdout, args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	case "init":
		if len(args) < 1 {
			printDeployUsage()
			os.Exit(1)
		}
		id := args[0]
		opts := deploy.InitOptions{OutDir: "."}
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--dry-run":
				opts.DryRun = true
			case "--force":
				opts.Force = true
			case "--out":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --out için dizin gerekli")
					os.Exit(1)
				}
				i++
				opts.OutDir = args[i]
			case "--module":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --module için değer gerekli")
					os.Exit(1)
				}
				i++
				opts.Module = args[i]
			case "--port":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --port için değer gerekli")
					os.Exit(1)
				}
				i++
				opts.Port = args[i]
			case "--main-package":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --main-package için değer gerekli")
					os.Exit(1)
				}
				i++
				opts.MainPackage = args[i]
			case "--health-path":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --health-path için değer gerekli")
					os.Exit(1)
				}
				i++
				opts.HealthPath = args[i]
			case "--api-base-url":
				if i+1 >= len(args) {
					fmt.Fprintln(os.Stderr, "deploy init: --api-base-url için değer gerekli")
					os.Exit(1)
				}
				i++
				opts.APIBaseURL = args[i]
			default:
				fmt.Fprintf(os.Stderr, "deploy init: bilinmeyen argüman: %s\n", args[i])
				printDeployUsage()
				os.Exit(1)
			}
		}
		if err := deploy.Init(id, opts); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if !opts.DryRun {
			fmt.Fprintf(os.Stdout, "deploy init %s tamam (çıktı dizini: %s)\n", id, opts.OutDir)
		}
	default:
		fmt.Fprintf(os.Stderr, "bilinmeyen deploy alt-komutu: %s\n", subcmd)
		printDeployUsage()
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// doctor
// ---------------------------------------------------------------------------

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
		tplTag := "custom"
		if config.IsKnownTemplate(p.Name) {
			tplTag = "built-in"
		}
		if p.EnvKey == "" {
			fmt.Printf("✗ %s [%s] -> env_key eksik\n", label, tplTag)
			failed = true
			continue
		}
		if os.Getenv(p.EnvKey) == "" {
			fmt.Printf("✗ %s [%s] -> env boş (%s)\n", label, tplTag, p.EnvKey)
			failed = true
			continue
		}
		fmt.Printf("✓ %s [%s] -> env hazır (%s)\n", label, tplTag, p.EnvKey)
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

	fmt.Printf("✓ mevcut template sayısı: %d (%s)\n", len(config.BuiltinTemplates), strings.Join(sortedTemplateNames(), ", "))

	if failed {
		fmt.Println("\nSonuç: bazı kontroller başarısız.")
		os.Exit(1)
	}
	fmt.Println("\nSonuç: tüm kritik kontroller başarılı.")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

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
	if tpl, ok := config.GetTemplate(providerName); ok {
		return tpl.DefaultModel
	}
	return ""
}

func suggestEnvKey(providerName string) string {
	if tpl, ok := config.GetTemplate(providerName); ok {
		return tpl.EnvKey
	}
	return strings.ToUpper(providerName) + "_API_KEY"
}

func sortedTemplateNames() []string {
	names := config.ListTemplateNames()
	sort.Strings(names)
	return names
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

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// ---------------------------------------------------------------------------
// usage text
// ---------------------------------------------------------------------------

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
  templates list                Mevcut provider template'lerini listele
  templates show <name>         Template detayını göster
  doctor                        Hızlı sistem sağlık kontrolü
  deploy list                   Deploy hedeflerini listele (railway, fly, vercel, …)
  deploy show <id>              Bir hedefin açıklaması ve üreteceği dosyalar
  deploy init <id> [flags]      Şablon dosyalarını repoya yazar (Dockerfile, fly.toml, …)
  version                       Sürüm bilgisi
  help                          Bu yardım mesajı`)
}

func printDeployUsage() {
	fmt.Println(`Kullanım:
  thy-case-llm deploy list
  thy-case-llm deploy show <id>
  thy-case-llm deploy init <id> [flags]

id: railway | fly | vercel (thy-case-llm deploy list)

init flags:
  --out <dir>           Çıktı kökü (varsayılan: .)
  --dry-run             Dosya yazma; içeriği stdout'a yaz
  --force               Var olan çıktı dosyalarının üzerine yaz
  --module <path>       go.mod module satırı yerine sabit modül adı
  --port <port>         Şablondaki PORT / internal_port
  --main-package <path> go build paket yolu (örn. ./cmd/api)
  --health-path <path>  Sağlık endpoint'i (örn. /health)
  --api-base-url <url>  vercel şablonunda rewrite hedefi (sonunda / olmasın)`)
}

func printProviderUsage() {
	fmt.Println(`Kullanım:
  thy-case-llm provider <alt-komut> [argümanlar]

Alt komutlar:
  add                           Yeni provider ekle
    --template <name>           Hazır template kullan (openai, gemini, anthropic)
    --name <name>               Provider adı
    --model <model>             Model adı
    --env-key <key>             API key ortam değişkeni adı
    --set-default               Varsayılan olarak ayarla
  list                          Kayıtlı provider'ları listele
  remove <name>                 Provider'ı kaldır
  set-default <name>            Varsayılan provider'ı değiştir
  validate                      Tüm provider'ları doğrula
  templates list                Mevcut template'leri listele
  templates show <name>         Template detayını göster
  doctor                        Provider + env + config kontrolü`)
}

func printTemplatesUsage() {
	fmt.Println(`Kullanım:
  thy-case-llm templates <alt-komut>

Alt komutlar:
  list                          Mevcut provider template'lerini listele
  show <name>                   Template detayını göster`)
}
