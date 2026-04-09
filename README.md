<p align="center">
  <img src="./assets/turkiye-header.svg?v=4" alt="Case Study Backend Side for THY" width="100%" />
</p>
<p align="center">
  <a href="https://pkg.go.dev/github.com/messivite/gosupabase">
    <img src="https://pkg.go.dev/badge/github.com/messivite/gosupabase.svg" alt="Go Reference: gosupabase" />
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white&style=for-the-badge" alt="Go Version" />
  </a>
  <a href="https://supabase.com/">
    <img src="https://img.shields.io/badge/Supabase-Ready-3ECF8E?logo=supabase&logoColor=white&style=for-the-badge" alt="Supabase" />
  </a>
  <img src="https://img.shields.io/badge/JWT-HS256%20%7C%20ES256-orange?style=for-the-badge" alt="JWT" />
  <img src="https://img.shields.io/badge/JWKS-Auto%20Fetch-2563eb?style=for-the-badge" alt="JWKS" />
  <img src="https://img.shields.io/badge/Dev-gosupabase%20dev-16a34a?style=for-the-badge" alt="Hot Reload" />
  <a href="https://github.com/messivite/go-thy-case-study-backend/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/messivite/go-thy-case-study-backend/ci.yml?branch=main&style=for-the-badge&label=CI" alt="CI" />
  </a>
  <a href="https://github.com/messivite/go-thy-case-study-backend/actions/workflows/release.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/messivite/go-thy-case-study-backend/release.yml?style=for-the-badge&label=Release" alt="Release" />
  </a>
  <a href="https://app.codecov.io/gh/messivite/go-thy-case-study-backend">
    <img src="https://img.shields.io/codecov/c/github/messivite/go-thy-case-study-backend?style=for-the-badge&label=Coverage" alt="Coverage" />
  </a>
</p>

# THY için Case Study Kapsamında Hazırlanan Backend Side Go Projesi

Supabase tabanlı kimlik doğrulama ve rol yönetimi kullanan, LLM sohbet akışlarını destekleyen Go backend uygulaması.

**Sürüm notları:** [CHANGELOG.md](CHANGELOG.md) - [RELEASE_NOTES.md](RELEASE_NOTES.md)

## Built With gosupabase

Bu proje, tarafımdan geliştirilen `gosupabase` kütüphanesi üzerine kuruludur. YAML tabanlı endpoint tanımları, JWT validate ve role-based access control (RBAC) katmanı `gosupabase` ile sağlanır ve yönetilir. Supabase ile yetenekleri çoğu fonksiyonu için tam uyumludur.


- GitHub: [github.com/messivite/gosupabase](https://github.com/messivite/gosupabase)
- Go Package: [pkg.go.dev/github.com/messivite/gosupabase](https://pkg.go.dev/github.com/messivite/gosupabase)

## Mimari Özeti

- **Auth:** Supabase access token (`Authorization: Bearer <jwt>`)
- **Roller:** `public.user_roles` -> hook -> JWT `claims.roles`
- **Route bazlı yetki:** `api.yaml` içindeki `roles: [...]`
- **Profil:** `public.profiles` ve `auth.users` arasında 1:1 ilişki
- **Chat persistence:** Varsayılan `supabase`; opsiyonel `memory`
- **LLM:** `providers.yaml` + environment variable anahtarları
- **Kota ve audit:** Supabase tabloları + trigger + RPC

## Proje Yapısı

```text
cmd/
  api/main.go              -> API sunucu giriş noktasi
  server/main.go           -> gosupabase dev uyumlu giriş noktasi
  thy-case-llm/main.go     -> LLM provider ve deploy yönetim CLI'i
internal/
  application/chat/        -> Use-case katmanı
  auth/                    -> JWT middleware ve context yardimcilari
  chat/                    -> HTTP handler katmanı
  config/                  -> Provider konfigürasyonu ve şablonlar
  deploy/                  -> deploy list/show/init şablonları
  domain/chat/             -> Domain modelleri, interface'ler, hatalar
  provider/                -> OpenAI/Gemini adapter'lari ve registry
  repo/                    -> Supabase ve memory repository implementasyonları
providers.yaml             -> LLM provider konfigürasyonu (non-secret)
api.yaml                   -> Endpoint tanımları ve rol kuralları
supabase/                  -> Migration, function ve Supabase config dosyalari
```

## Provider Konfigurasyonu

Provider metadata ve gizli anahtarlar ayrıdır:

| Dosya | İçerik | Git'e eklenir? |
|---|---|---|
| `providers.yaml` | Provider adı, model, env key referansı | Evet |
| `.env` | Gerçek API key değerleri | Hayır |

Örnek `providers.yaml`:

```yaml
default: openai
providers:
  - name: openai
    model: gpt-4o
    env_key: OPENAI_API_KEY
  - name: gemini
    model: gemini-2.5-flash
    env_key: GEMINI_API_KEY
```

Uygulama açılışında `providers.yaml` okunur. Env key değeri bulunamayan provider kaydı atlanır ve log uyarısı üretilir.

## thy-case-llm CLI

Yardım:

```bash
go run ./cmd/thy-case-llm help
```

Global kurulum:

```bash
go install ./cmd/thy-case-llm
```

Tüm komutları derleyerek kurmak:

```bash
git clone https://github.com/messivite/go-thy-case-study-backend.git
cd go-thy-case-study-backend
go install ./cmd/...
```

PATH'e ekleme (gerekirse):

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
thy-case-llm version
```

### Komutlar

| Komut | Açıklama |
|---|---|
| `provider add` | Yeni provider ekler (`--template` destekler) |
| `provider list` | Kayıtlı provider'ları listeler |
| `provider remove <name>` | Provider kaydını siler |
| `provider set-default <name>` | Varsayılan provider'ı değiştirir |
| `provider validate` | Env key doğrulaması yapar |
| `templates list` | Yerleşik provider şablonlarını listeler |
| `templates show <name>` | Şablon detayını gösterir |
| `doctor` | Hızlı sistem sağlık kontrolü |
| `deploy list` | Desteklenen deploy hedeflerini listeler |
| `deploy show <id>` | Hedef ve yazılacak dosya detayını gösterir |
| `deploy init <id>` | Şablon dosyalarını repoya yazar |

### Kullanım Örnekleri

```bash
thy-case-llm provider list
thy-case-llm templates list
thy-case-llm templates show openai
thy-case-llm provider add --template openai --set-default
thy-case-llm provider validate
thy-case-llm doctor
```

## Environment

`.env.example` içindeki temel değişkenler:

- `PORT` (varsayılan `8081`)
- `CHAT_PERSISTENCE` (`supabase` veya `memory`)
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`
- `SUPABASE_JWT_VALIDATION_MODE` (`auto` önerilir)
- `SUPABASE_ROLE_CLAIM_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `PROVIDERS_CONFIG` (varsayılan `providers.yaml`)
- `OBSERVABILITY_LOG_FILE` (opsiyonel)
- `OTEL_EXPORTER_OTLP_ENDPOINT` (opsiyonel)

### CHAT_PERSISTENCE

| Değer | Davranış |
|---|---|
| `supabase` veya boş | Sohbet verisi Supabase `chat_sessions` / `chat_messages` tablolarına yazılır |
| `memory` | Veri sadece process RAM'inde tutulur; process kapanınca silinir |

Notlar:

- `CHAT_PERSISTENCE=supabase` iken `SUPABASE_URL` veya `SUPABASE_SERVICE_ROLE_KEY` yoksa uygulama memory moduna fallback eder.
- `cmd/api` ve `cmd/server` açılışında lokal `.env` dosyasını yükler (`godotenv.Overload`); dosya değerleri stale shell export değerlerinin üzerine yazılır.
- Üretim ortamında `.env` dosyası yerine platform environment değişkenleri kullanılmalıdır.

### Observability

`internal/observability` paketi JSON line formatında log üretir. `OBSERVABILITY_LOG_FILE` ayarlanırsa aynı loglar dosyaya append edilir.

### OpenTelemetry

Bu repoda minimal HTTP trace entegrasyonu vardır. `OTEL_EXPORTER_OTLP_ENDPOINT` tanımlı değilse tracing devreye girmez.

Örnek:

```bash
./otelcol --config=/ABSOLUTE/PATH/thy-case-study-backend/otel/collector.yaml
```

## Endpointler

| Metot | Endpoint | Auth | Açıklama |
|---|---|---|---|
| `GET` | `/health` veya `/api/health` | Hayır | Health check (`OK`) |
| `GET` | `/api/me` | Evet | JWT'den user bilgisi |
| `GET` | `/api/providers` | Evet | Aktif provider listesi ve default bilgi |
| `POST` | `/api/chats` | Evet | Yeni sohbet oluşturur |
| `GET` | `/api/chats` | Evet | Sohbet listesini döner |
| `GET` | `/api/chats/{chatID}` | Evet | Sohbet ve mesaj detaylarını döner |
| `POST` | `/api/chats/{chatID}/messages` | Evet | Non-stream mesaj gönderir |
| `POST` | `/api/chats/{chatID}/stream` | Evet | SSE stream mesaj gönderir |

## Auth ve Rol Akışı

1. Kullanıcı Supabase ile giriş yapar ve access token alır.
2. Hook mekanizması `user_roles` tablosundan rollerin JWT claim'lerine yazılmasını sağlar.
3. API token'i doğrular.
4. Endpoint bazlı rol kuralları `api.yaml` üzerinden uygulanır.

Rol değişikliği sonrasında yeni token alınmalıdır.

## Rol Atama

Rol ekleme:

```sql
insert into public.user_roles (user_id, role)
values ('USER_UUID_HERE'::uuid, 'editor')
on conflict (user_id, role) do nothing;
```

Rol silme:

```sql
delete from public.user_roles
where user_id = 'USER_UUID_HERE'::uuid
  and role = 'admin';
```

## PostgreSQL (Supabase) Veri Modeli

- `auth.users` -> Kimlik kayıtları
- `public.profiles` -> Kullanıcı profil verisi (1:1)
- `public.user_roles` -> Çoklu rol ilişkisi
- `public.chat_sessions`, `public.chat_messages` -> Sohbet verisi
- `public.llm_interaction_log` -> LLM audit ve usage logu
- `public.llm_quota_defaults`, `public.user_llm_usage_quota` -> Kota konfigürasyonu

## Supabase Kurulum

Projeyi linkleme:

```bash
npx supabase link --project-ref <PROJECT_REF>
```

Migration uygulama:

```bash
npx supabase db push --linked
```

Function deploy:

```bash
npx supabase functions deploy register-push-token --project-ref <PROJECT_REF>
```

## CI ve Coverage

GitHub Actions pipeline'i `build`, `test` ve `vet` adımlarını çalıştırır. Coverage çıktıları Codecov'a gönderilir.

## Yerelde Çalıştırma

```bash
gosupabase dev
```

veya:

```bash
go run ./cmd/api
```

## Faz 0 Test (curl)

```bash
TOKEN="<ACCESS_TOKEN>"
```

Sohbet oluşturma:

```bash
curl -X POST "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"ilk test chat session","provider":"gemini","model":"gemini-2.5-flash"}'
```

Mesaj gönderme (non-stream):

```bash
curl -X POST "http://localhost:8081/api/chats/<CHAT_ID>/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider":"openai",
    "model":"gpt-4.1-mini",
    "messages":[
      {"role":"user","content":"Merhaba"}
    ]
  }'
```

Mesaj gönderme (stream):

```bash
curl -N -X POST "http://localhost:8081/api/chats/<CHAT_ID>/stream" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider":"gemini",
    "model":"gemini-2.5-flash",
    "messages":[
      {"role":"user","content":"Kısa bir selamlama yaz"}
    ]
  }'
```

## Faz 3 Durumu

### Tamamlananlar

- `thy-case-llm deploy` komutları (railway/fly/vercel şablonları)
- LLM interaction audit kaydı (`llm_interaction_log`)
- Token kota modeli (`llm_quota_defaults`, `user_llm_usage_quota`)
- Profil oluşumunda quota satırı üreten trigger
- Kota aşımında tutarlı HTTP 429 hata kodları

### Planlanan

- Self-hosted / özel endpoint tanımlarının provider konfigürasyonuna eklenmesi

## Deploy

CLI v0.3.0+ ile deploy şablonları üretilir:

```bash
thy-case-llm deploy list
thy-case-llm deploy show railway
thy-case-llm deploy init railway --dry-run
thy-case-llm deploy init railway
```

Desteklenen hedefler:

| id | Uretilen dosyalar | Not |
|---|---|---|
| `railway` | `Dockerfile`, `railway.toml` | Varsayılan health path `/health` |
| `fly` | `Dockerfile`, `fly.toml` | Benzer Docker tabanlı kurulum |
| `vercel` | `vercel.json`, `deploy/VERCEL.md` | Rewrite temelli yonlendirme senaryosu |

Yaygin flag'ler: `--dry-run`, `--force`, `--out`, `--port`, `--main-package`, `--health-path`, `--api-base-url`, `--module`.
<p align="center">
  <img src="./assets/turkiye-header.svg?v=4" alt="THY Case Study Backend" width="100%" />
</p>
<p align="center">
  <a href="https://pkg.go.dev/github.com/messivite/gosupabase">
    <img src="https://pkg.go.dev/badge/github.com/messivite/gosupabase.svg" alt="Go Reference: gosupabase" />
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white&style=for-the-badge" alt="Go Version" />
  </a>
  <a href="https://supabase.com/">
    <img src="https://img.shields.io/badge/Supabase-Ready-3ECF8E?logo=supabase&logoColor=white&style=for-the-badge" alt="Supabase" />
  </a>
  <img src="https://img.shields.io/badge/JWT-HS256%20%7C%20ES256-orange?style=for-the-badge" alt="JWT" />
  <img src="https://img.shields.io/badge/JWKS-Auto%20Fetch-2563eb?style=for-the-badge" alt="JWKS" />
  <img src="https://img.shields.io/badge/Dev-gosupabase%20dev-16a34a?style=for-the-badge" alt="Hot Reload" />
  <a href="https://github.com/messivite/go-thy-case-study-backend/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/messivite/go-thy-case-study-backend/ci.yml?branch=main&style=for-the-badge&label=CI" alt="CI" />
  </a>
  <a href="https://github.com/messivite/go-thy-case-study-backend/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/messivite/go-thy-case-study-backend/ci.yml?branch=main&style=for-the-badge&label=Tests" alt="Tests" />
  </a>
  <a href="https://github.com/messivite/go-thy-case-study-backend/actions/workflows/release.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/messivite/go-thy-case-study-backend/release.yml?style=for-the-badge&label=Release" alt="Release" />
  </a>
  <a href="https://app.codecov.io/gh/messivite/go-thy-case-study-backend">
    <img src="https://img.shields.io/codecov/c/github/messivite/go-thy-case-study-backend?style=for-the-badge&label=Coverage" alt="Coverage" />
  </a>
  <a href="https://github.com/messivite/gosupabase/releases">
    <img src="https://img.shields.io/github/v/release/messivite/gosupabase?label=gosupabase%20release" alt="gosupabase release" />
  </a>
  <a href="https://github.com/messivite/gosupabase/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/messivite/gosupabase" alt="License" />
  </a>
  <img src="https://img.shields.io/badge/Router-chi-34495E" alt="Router" />
  <img src="https://img.shields.io/badge/Codegen-YAML%20Driven-7D3C98" alt="Codegen" />
  <img src="https://img.shields.io/badge/YAML-First-1ABC9C" alt="YAML First" />
  <a href="https://github.com/messivite/go-thy-case-study-backend/commits/main">
    <img src="https://img.shields.io/github/last-commit/messivite/go-thy-case-study-backend" alt="Last Commit" />
  </a>
</p>

# Thy Case Study Backend

THY case study için yazdığım Go backend. Auth Supabase, roller JWT claim üzerinden; route’lar ve guard’lar `gosupabase` ile.

**Sürüm notları:** [CHANGELOG.md](CHANGELOG.md) · [RELEASE_NOTES.md](RELEASE_NOTES.md) (tag / GitHub Release akışı)

## Built With gosupabase

Bu proje, tarafımdan geliştirilen `gosupabase` kütüphanesi üzerine inşa edilmiştir; YAML tabanlı endpoint tanımları, JWT doğrulama ve role-based access control (RBAC) katmanı bu kütüphane üzerinden sağlanır.

- GitHub: [github.com/messivite/gosupabase](https://github.com/messivite/gosupabase)
- Go Package: [pkg.go.dev/github.com/messivite/gosupabase](https://pkg.go.dev/github.com/messivite/gosupabase)

## Mimari Özeti

- **Auth:** Supabase access token (`Authorization: Bearer <jwt>`)
- **Roller:** `public.user_roles` → hook → JWT’de `claims.roles`
- **Route rolleri:** `api.yaml` içindeki `roles: [...]`
- **Profil:** `public.profiles`, `auth.users` ile 1:1
- **Chat:** Varsayılan Supabase Postgres (REST + service role); istersen `CHAT_PERSISTENCE=memory` ile sadece RAM
- **DB:** Supabase Postgres (auth + `profiles` + `user_roles` + chat tabloları, RLS)
- **LLM:** `providers.yaml` + `.env`, yönetim için `thy-case-llm` (`deploy` ile Railway/Fly/Vercel şablonları)

## Proje Yapısı

```
cmd/
  api/main.go              → API sunucu giriş noktası
  server/main.go           → gosupabase dev uyumlu giriş noktası
  thy-case-llm/main.go     → LLM provider yönetim CLI'ı
internal/
  deploy/                  → deploy list/show/init (Railway, Fly, Vercel şablonları, go:embed)
  domain/chat/             → Domain modelleri, interface'ler, error'lar
    models.go              → Role, ChatMessage, ChatSession, StreamEvent, ProviderRequest/Response
    provider.go            → LLMProvider interface (Complete + Stream)
    repository.go          → Repository interface
    errors.go              → Domain hata tanımları
  application/chat/        → Use-case katmanı (iş kuralları)
    usecase.go             → CreateSession, SendMessage, StreamMessage, ...
  provider/                → LLM provider adapter'ları
    registry.go            → Provider registry (metadata, default, list)
    openai.go              → OpenAI adapter
    gemini.go              → Gemini adapter
    helpers.go             → Ortak yardımcı fonksiyonlar
  config/                  → Konfig + built-in LLM şablonları
    provider.go            → providers.yaml
    templates.go           → CLI template registry (openai, gemini, …)
  chat/                    → HTTP handler katmanı
    handler.go             → REST + SSE endpoint'leri
  repo/                    → Repository implementasyonları
    memory_repository.go   → RAM (CHAT_PERSISTENCE=memory)
    supabase_repository.go → Postgres (Supabase REST, varsayılan)
  auth/                    → JWT doğrulama
    auth.go                → Middleware, context helpers
    supabase_adapter.go    → Supabase JWT/JWKS adapter
  app/
    server.go              → Router + middleware bağlama
providers.yaml             → LLM provider konfigürasyonu (non-secret)
api.yaml                   → Endpoint listesi + üstte chat sözleşme notları (body/response şekilleri)
supabase/                  → Supabase CLI config, functions, migrations
```

## Provider Konfigürasyonu

Anahtarlar repoda durmasın diye ikiye ayırdım:

| Dosya | İçerik | Git'e eklenir? |
|-------|--------|----------------|
| `providers.yaml` | Provider adı, model, env key referansı | Evet |
| `.env` | API anahtarları (`OPENAI_API_KEY`, `GEMINI_API_KEY`, ...) | Hayır |

**providers.yaml örneği:**

```yaml
default: openai
providers:
  - name: openai
    model: gpt-4o
    env_key: OPENAI_API_KEY
  - name: gemini
    model: gemini-2.5-flash
    env_key: GEMINI_API_KEY
```

Sunucu açılınca `providers.yaml` okunuyor; env’de key yoksa o provider atlanıyor, log’a uyarı düşüyor.

## thy-case-llm CLI

Provider’ları terminalden yönetmek için küçük bir araç.

```bash
go run ./cmd/thy-case-llm help
```

Global komut olarak kullanmak için:

```bash
cd /Users/kullanıcıAdınız/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
```

**Repoyu klonlayıp tüm çalıştırılabilirleri `go install` ile kurmak:** modül kökünde (Go 1.16+; bu repo 1.25) şu yeterli; `go mod download` ayrıca gerekmez, derleme sırasında bağımlılıklar iner:

```bash
git clone https://github.com/messivite/go-thy-case-study-backend.git
cd go-thy-case-study-backend
go install ./cmd/...
```

`$(go env GOPATH)/bin` içine şunlar yazılır: **`api`**, **`server`**, **`thy-case-llm`**. Bu dizin `PATH`’te değilse:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
thy-case-llm version
# api ve server doğrudan sunucu başlatır; çalıştırmak için repo kökünde .env ile:
#   api
#   server
```

Çalışma zamanı için ayrıca `.env` / `providers.yaml` / Supabase ayarları gerekir; bunlar `go install` ile gelmez.

### Önemli not (güncel binary)

`templates list`, `templates show`, `provider add --template …` Faz 2’den beri var. Eski `go install` binary’si kalırsa şunu görürsün:

```text
unknown command: templates
```

`thy-case-llm help` içinde de `templates` yoksa PATH’teki binary eski demektir.

**Ne yapayım:** Projede tekrar `go install`, `$(go env GOPATH)/bin` PATH’te olsun:

```bash
cd /Users/kullanıcıAdınız/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
which thy-case-llm
thy-case-llm version
thy-case-llm templates list
```

Install istemezsen: `go run ./cmd/thy-case-llm templates list`.

### Komutlar

| Komut | Açıklama |
|-------|----------|
| `provider add` | Yeni provider ekle (interaktif veya `--name`, `--model`, `--env-key` flag'leri ile; `--template <name>` ile hazır şablon) |
| `provider list` | Kayıtlı provider'ları listele (varsayılan, model, env durumu) |
| `provider remove <name>` | Provider'ı kaldır |
| `provider set-default <name>` | Varsayılan provider'ı değiştir |
| `provider validate` | Tüm provider'ların env key kontrolünü yap |
| `templates list` | Yerleşik provider şablonlarını listele (openai, gemini, anthropic, …) |
| `templates show <name>` | Bir şablonun detayını göster (model, env key, base URL) |
| `doctor` | Provider + env + config için hızlı sağlık kontrolü |
| `deploy list` | Üretilebilir deploy hedeflerini listele (`railway`, `fly`, `vercel`) |
| `deploy show <id>` | Hedef açıklaması ve yazılacak dosya yolları |
| `deploy init <id>` | Şablonları repoya yazar (`Dockerfile`, `fly.toml`, örnek `vercel.json`, …) |

Deploy komutunun tam açıklaması ve örnekler [en alttaki Deploy bölümünde](#deploy).

### Kullanım Örnekleri

```bash
# Provider listele
thy-case-llm provider list

# Hazır şablonları listele ve detay gör
thy-case-llm templates list
thy-case-llm templates show openai

# Yeni provider ekle (interaktif)
thy-case-llm provider add

# Hazır şablonla ekle (ör. OpenAI)
thy-case-llm provider add --template openai --set-default

# Yeni provider ekle (flag ile)
thy-case-llm provider add --name openai --model gpt-4o --env-key OPENAI_API_KEY

# Varsayılan provider'ı gemini yap
thy-case-llm provider set-default gemini

# Konfigürasyon doğrula
thy-case-llm provider validate

# Hızlı sağlık kontrolü
thy-case-llm doctor

# Provider kaldır
thy-case-llm provider remove gemini
```

### Olası Hata ve Çözüm

`zsh: command not found: thy-case-llm` çıkarsa:

1) Binary'yi kurun:

```bash
cd /Users/kullanıcıAdınız/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
```

2) Go bin path'i `PATH` içinde değilse ekleyin:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

3) Ya da direkt:

```bash
go run ./cmd/thy-case-llm doctor
```

## Environment

`.env.example`:

- `PORT` (default `8081`)
- `CHAT_PERSISTENCE` — ayrıntı ve sorun giderme: aşağıdaki **[CHAT_PERSISTENCE](#chat_persistence)** bölümü.
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`
- `SUPABASE_JWT_VALIDATION_MODE` (`auto` önerilir)
- `SUPABASE_ROLE_CLAIM_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `PROVIDERS_CONFIG` (varsayılan `providers.yaml`, opsiyonel)
- `OBSERVABILITY_LOG_FILE` (opsiyonel) — yapılandırılmış logları **JSON Lines** olarak bu dosyaya da yazar; stdout davranışı aynı kalır. Örn. `./logs/app.jsonl` veya `/var/log/thy-api.jsonl`. Boş bırakılınca sadece process stdout (chi logger ayrı). Loki, Vector, CloudWatch agent, `tail -f` ile izlenebilir.
- `OTEL_EXPORTER_OTLP_ENDPOINT` (opsiyonel) — OpenTelemetry trace export (OTLP HTTP); ayrıntı aşağıda.

### CHAT_PERSISTENCE

| Değer | Davranış |
|-------|----------|
| `supabase` veya **boş** (varsayılan) | Sohbet **`chat_sessions` / `chat_messages`** üzerinden Supabase REST + service role. Kota ve `llm_interaction_log` da bu modda anlamlıdır. |
| `memory` | Veri **yalnızca süreç RAM’inde** (`MemoryRepository`); sunucuyu kapatınca silinir. Kota katmanı **stub** (`MemoryQuotaRepository`, pratikte limit uygulanmaz). |

**Supabase’e düşmeme:** `CHAT_PERSISTENCE=supabase` iken `SUPABASE_URL` veya `SUPABASE_SERVICE_ROLE_KEY` boşsa kod **otomatik memory’ye** geçer ve log’a uyarı yazar.

**`.env` (yerel):** `cmd/api` ve `cmd/server` açılışta **bulduğu ilk `.env`** dosyasını yükler (cwd’den üst dizinlere doğru arar, `internal/dotenv`). [`Overload`](https://github.com/joho/godotenv) kullanılır: **dosyadaki değerler shell’de kalmış eski `export`’ların üzerine yazılır** (ör. `CHAT_PERSISTENCE=memory` unutulduysa `.env` içindeki `supabase` geçerli olur). **Üretim imajına `.env koymayın**; yalnızca platform env kullanın. Tek seferlik `CHAT_PERSISTENCE=memory go run …` denemesi için `.env` satırını geçici yorum satırı yap veya `env -u CHAT_PERSISTENCE go run ./cmd/api`.

**İsteğe bağlı:** Hâlâ `source .env` veya `VAR=değer go run …` kullanabilirsin.

**Doğrulama:** Açılış log’u `chat persistence: supabase (postgres)` veya `chat persistence: memory (in-process)` — hangi modda olduğunu buradan teyit et.

### Observability (dosyaya log)

`internal/observability` içindeki `Info` / `Warn` / `Error` ve `LLM*` helper’ları tek satırlık JSON üretir (`ts`, `level`, `event`, `fields`, …). `OBSERVABILITY_LOG_FILE` set edilirse **aynı satırlar dosyaya append** edilir; process sonunda dosya `CloseFileLog` ile kapatılır (`cmd/api` ve `cmd/server`).

Örnek satır:

```json
{"ts":"2026-04-09T12:00:00.123456789Z","level":"info","event":"llm.request","fields":{"provider":"openai","model":"gpt-4.1-mini","user_id":"…","session_id":"…"}}
```

Üretimde dosya boyutu için logrotate / sidecar veya merkezi toplayıcı kullan; uygulama içinde rotation yok (bilinçli sade tutuş).

### OpenTelemetry (trace)

**Zorluk:** Orta — bu repoda **minimal** kurulum var: gelen HTTP istekleri için **server span** (chi router + `otelhttp`), export **OTLP/HTTP**. Env yoksa **hiçbir şey çalışmaz** (sıfır ek yük).

| Env | Anlam |
|-----|--------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` veya `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Örn. `http://localhost:4318` (Collector HTTP). Tanımlı değilse tracing kapalı. |
| `OTEL_SERVICE_NAME` | Varsayılan `thy-case-study-api`. |

Örnek Collector config bu repoda: **`otel/collector.yaml`**. Binary nerede olursa olsun **config’e tam yol** ver:

```bash
./otelcol --config=/ABSOLUTE/PATH/thy-case-study-backend/otel/collector.yaml
```

(`config.yaml: no such file` hatası, komutu çalıştırdığın dizinde dosya arandığı içindir; `--config=` ile proje içindeki yolu göster.)

#### OpenTelemetry Collector kurulumu (yerel test)

1. **İndir** — Resmi sürümler [OpenTelemetry Collector releases](https://github.com/open-telemetry/opentelemetry-collector-releases/releases) sayfasında. Bu repodaki örnek config **Core** dağıtımı (`otelcol`) ile uyumludur; macOS için `otelcol_*_darwin_arm64.tar.gz` veya `*_amd64` uygun olanı seç. Daha fazla exporter/receiver gerekiyorsa aynı sayfadan **[otelcol-contrib](https://github.com/open-telemetry/opentelemetry-collector-releases/releases)** varlığını da kullanabilirsin. Kurulum özeti: [Collector installation](https://opentelemetry.io/docs/collector/installation/).
2. **Arşivi aç**, `otelcol` binary’sini istediğin klasöre koy. İlk çalıştırmada macOS “bilinmeyen geliştirici” diyebilir: Finder’da sağ tık → **Aç** veya `xattr -dr com.apple.quarantine ./otelcol`.
3. **Collector’ı ayrı terminalde çalıştır** (pencere açık kalsın):

   ```bash
   ./otelcol --config=/ABSOLUTE/PATH/thy-case-study-backend/otel/collector.yaml
   ```

   Log’da `Starting HTTP server` … `endpoint: "[::]:4318"` görünmeli.
4. **API tarafı** — `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318` (ve isteğe bağlı `OTEL_SERVICE_NAME`) tanımlı olsun; repo kökündeki `.env` içine yazabilirsin (`cmd/api` açılışta yükler) veya export / IDE env kullan.
5. **Test** — `GET /health` ve `GET /api/health` trace üretmez; `GET /api/providers` veya `GET /api/chats` gibi korumalı bir endpoint’e JWT ile istek at. Trace çıktısı **Collector’ın çalıştığı terminalde** (`debug` exporter) görünür.

**Örnek ekran görüntüsü** — Collector’da gelen span (bu projeden `GET /api/chats`, `debug` exporter çıktısı):

![OpenTelemetry Collector debug exporter ile gelen trace örneği](https://i.ibb.co/VWrfvghD/Screenshot-2026-04-09-at-02-22-32.png)

*(Aynı görüntü doğrudan link: [Screenshot-2026-04-09-at-02-22-32.png](https://i.ibb.co/VWrfvghD/Screenshot-2026-04-09-at-02-22-32.png))*

Üretimde Collector’a Jaeger, Grafana Tempo, vendor OTLP uçları gibi **exporter** eklenir; yerelde sadece `debug` ile konsolda doğrulamak yeterli.

**Sonraki seviye (yapılmadı):** LLM çağrıları için `usecase` içinde `otel.Tracer` ile child span, metrics, log→trace bağlama — ihtiyaç oldukça eklenebilir.

## Endpointler

| Metot | Endpoint | Auth | Açıklama |
|-------|----------|------|----------|
| `GET` | `/health` veya `/api/health` | Hayır | Aynı yanıt (`OK`). Probe için `/health`; `baseUrl` `…/api` ise path sadece `/health` — tam path `/api/health` olur. `baseUrl` zaten `…/api` iken path’e bir daha `/api/health` ekleme (çift `/api` 404). |
| `GET` | `/api/me` | Evet | JWT'deki kullanıcı bilgisi |
| `GET` | `/api/providers` | Evet | Aktif LLM provider'ları (default bilgisi dahil) |
| `POST` | `/api/chats` | Evet | Yeni sohbet; gövdede isteğe bağlı `provider`, `model` (`providers.yaml` / `GET /api/providers`). Cevap: `id`, `provider`, `model` (session default’ları) |
| `GET` | `/api/chats` | Evet | Sohbet listesi; her öğede `provider` / `model` (son LLM turu veya session default) |
| `GET` | `/api/chats/{chatID}` | Evet | Sohbet + mesajlar; her asistan satırında `provider`/`model` (yeni mesajlardan); kökte özet = son dolu asistan veya session |
| `POST` | `/api/chats/{chatID}/messages` | Evet | Mesaj gönder (non-stream) |
| `POST` | `/api/chats/{chatID}/stream` | Evet | Mesaj gönder (SSE stream) |

## Auth + Rol Akışı

1. Login → Supabase access token.
2. Hook `user_roles`’a bakıp `claims.roles`’u token’a yazıyor.
3. API JWT’yi `gosupabase` ile doğruluyor.
4. `api.yaml`’da `roles` varsa token’daki rollerle eşleşmiyorsa `403`.

Rol değişince eski token yetmez; yenile veya tekrar login.

## Rol Atama

Bir kullanıcıya rol eklemek:

```sql
insert into public.user_roles (user_id, role)
values ('USER_UUID_HERE'::uuid, 'editor')
on conflict (user_id, role) do nothing;
```

Admin rolü vermek:

```sql
insert into public.user_roles (user_id, role)
values ('USER_UUID_HERE'::uuid, 'admin')
on conflict (user_id, role) do nothing;
```

Admin rolünü kaldırmak:

```sql
delete from public.user_roles
where user_id = 'USER_UUID_HERE'::uuid
  and role = 'admin';
```

Kullanıcının mevcut rollerini görmek:

```sql
select user_id, role
from public.user_roles
where user_id = 'USER_UUID_HERE'::uuid;
```

Mevcut rolü güncellemek (ör. `admin` -> `editor`):

```sql
update public.user_roles
set role = 'editor'
where user_id = 'USER_UUID_HERE'::uuid
  and role = 'admin';
```

## Rol Değişimi Nasıl Tetikleniyor?

1. `user_roles`’a insert/update/delete.
2. `trg_user_roles_sync_auth_metadata` çalışıyor.
3. `auth.users.raw_app_meta_data.roles` güncelleniyor.
4. Yeni token’da hook yine `user_roles`’tan okuyup `claims.roles` yazıyor.

Eski token’ın içi değişmez; rol değişince mutlaka yeni token lazım.

## Profiles Akışı

- User oluşunca trigger ile `profiles` satırı açılıyor.
- `is_anonymous` auth’taki flag’ten geliyor.
- `display_name`, `avatar_url` vs. app’ten update.

## PostgreSQL (Supabase) Veri Modeli

- `auth.users` — kimlik
- `public.profiles` — profil, user ile 1:1
- `public.user_roles` — çoklu rol
- Chat için `chat_sessions` / `chat_messages` (migration’larda)
- RLS Postgres’te; API tarafında JWT

## Supabase Kurulum Komutları

Projeyi linkle:

```bash
npx supabase link --project-ref <PROJECT_REF>
```

Migrationları uygula:

```bash
npx supabase db push --linked
```

Edge function deploy:

```bash
npx supabase functions deploy register-push-token --project-ref <PROJECT_REF>
```

## CI ve Codecov

GitHub Actions’ta build + test + vet; coverage `coverage.out` ile Codecov’a gidiyor. PR’da check’leri zorunlu tutabilirsin (codecov/patch vs.). İlk kurulum PR’ı: [trigger CI checkes #1](https://github.com/messivite/go-thy-case-study-backend/pull/1)

## Yerelde Çalıştırma

```bash
gosupabase dev
```

veya:

```bash
go run ./cmd/api
```

## Faz 0 Test (curl)

`TOKEN` yerine Supabase access token verin.

```bash
TOKEN="<ACCESS_TOKEN>"
```

**Sohbet oluştur — `POST /api/chats`**

İstek gövdesi (JSON):

| Alan | Zorunlu | Açıklama |
|------|---------|----------|
| `title` | Hayır | Sohbet başlığı (boş string olabilir) |
| `provider` | Hayır | Örn. `openai`, `gemini`. Yoksa `providers.yaml` içindeki `default` provider kullanılır ve `chat_sessions.default_provider` olarak kaydedilir. |
| `model` | Hayır | API model id (örn. `gemini-2.5-flash`). Yoksa seçilen provider’ın varsayılan model’i kullanılır. |

Bu `provider` / `model` değerleri oturum açılırken veritabanına yazılır; henüz mesaj yokken `GET /api/chats/{chatID}` cevabındaki kök `provider` / `model` alanları da buradan gelir (sonra mesajlar geldikçe öncelik mesajlardaki / son tur metadatasına göre güncellenir).

**201 cevap gövdesi:** `id` (UUID string), `provider`, `model` — kaydedilen session default’ları.

Sadece başlık:

```bash
curl -X POST "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"ilk test chat session"}'
```

Başlık + açık provider/model (Postman’da **Body → raw → JSON**; **Authorization’da yalnızca Bearer token** kullan — hem header hem ayrıca “Bearer Token” auth’u işaretlemek token’ı iki kez gönderebilir):

```bash
curl -X POST "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"ilk test chat session","provider":"gemini","model":"gemini-2.5-flash"}'
```

Hangi isimlerin geçerli olduğunu görmek için: `GET /api/providers` veya repodaki `providers.yaml`.

Chat listele (her satırda `provider` / `model`: son başarılı LLM turu veya oturum açılışında kayıtlı default):

```bash
curl "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN"
```

Chat detay:

```bash
curl "http://localhost:8081/api/chats/<CHAT_ID>" \
  -H "Authorization: Bearer $TOKEN"
```

Kök `provider`/`model` son anlamlı asistan cevabının özetidir. Her mesajda asistan satırları için `provider`/`model` vardır (eski kayıtlar migration öncesi boş/omit). `chat_messages.provider`, `chat_messages.model` ve isteğe bağlı session özeti için migration’ları uygula (`supabase db push`).

Mesaj gönder (non-stream):

```bash
curl -X POST "http://localhost:8081/api/chats/<CHAT_ID>/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider":"openai",
    "model":"gpt-4.1-mini",
    "messages":[
      {"role":"user","content":"Merhaba, nasilsin?"}
    ]
  }'
```

Mesaj gönder (SSE stream):

```bash
curl -N -X POST "http://localhost:8081/api/chats/<CHAT_ID>/stream" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "provider":"gemini",
    "model":"gemini-2.5-flash",
    "messages":[
      {"role":"user","content":"Bana kısa bir selamlama yaz"}
    ]
  }'
```

## Notlar

- `supabase/migrations` içinde bazı dosyalar eski geçişlerden kaldı, silmiyorum referans için.
- Rol kaynağı: `public.user_roles`.
- Yeni iş için `public.users` değil, `public.profiles`.
- **Faz 1:** Domain / application / infra ayrımı, `providers.yaml` + `thy-case-llm`, yeni LLM için adapter + registry.
- **Faz 2:** Template CLI (`templates list/show`), gerçek OpenAI/Gemini çağrıları, chat’i Postgres’e yazma, usage normalize, log tarafı, retry/timeout, provider hatalarını HTTP’ye map etme.

### Faz 3

`thy-case-llm deploy` şablonları ve canlı örnek API (Railway) tamam.

#### Self-hosted / özel endpoint (henüz yapılmadı)

Kendi makinemde veya şirket gateway’inde model çalıştırırsam şu an kod sadece OpenAI’nin sabit URL’ine gidiyor. Faz 3’te base URL’yi env veya `providers.yaml`’dan verebilir hale getirmek istiyorum (vLLM, LiteLLM proxy vs.). Gerekirse ek header. CLI’da da bu endpoint’i tanımlama. “Sadece farklı model id” ile “farklı host” ayrımını dokümanda net yazarım.

#### Token kotası (tamam)

- **`llm_quota_defaults`** (singleton): varsayılan `default_daily_tokens` (100k), `default_weekly_tokens` (500k).
- **`user_llm_usage_quota`** (PK = `user_id`): kullanıcıya özel günlük/haftalık limit + **`quota_bypass`** (admin override).
- **Profil trigger:** `AFTER INSERT ON public.profiles` ile kota satırı otomatik; mevcut kullanıcılar backfill.
- **Kontrol:** `SendMessage` / `StreamMessage` girişinde kota sorgulanır; aşılmışsa **HTTP 429** + `llm_quota_daily_exceeded` veya `llm_quota_weekly_exceeded`.

#### LLM etkileşim günlüğü (tamam)

`public.llm_interaction_log` + `chat_messages` AFTER INSERT trigger.

- Kullanıcı mesajı LLM çağrısından **önce** kaydedilir; trigger `pending` audit satırı oluşturur.
- Başarıda assistant INSERT trigger `ok` yapar; Go API token sayılarını `llm_set_usage_for_user_message` RPC ile yazar.
- LLM hatası olursa Go API `llm_fail_pending_for_user_message` RPC ile `error` + `error_code` / `provider_http_status` yazar.
- RLS yok; inceleme Supabase SQL / service role.

## Deploy

**thy-case-llm** ile üretim dosyalarını repoya yazdırmak: `thy-case-llm deploy` (CLI **v0.3.0+**; `thy-case-llm version`, eskiyse `go install ./cmd/thy-case-llm`).

API süreci **`cmd/api`**; dinlediği port varsayılan **`8081`** (`PORT` env). Docker build **her zaman repo kökünden** yapılmalı (`docker build -f Dockerfile .`).

### Hedefler (`deploy list`)

| `id` | Üretilen dosyalar | Not |
|------|-------------------|-----|
| `railway` | `Dockerfile`, `railway.toml` | Railway’de health check path şablonda `/health` |
| `fly` | `Dockerfile`, `fly.toml` | Dockerfile `railway` ile aynı şablondan |
| `vercel` | `vercel.json`, `deploy/VERCEL.md` | Go API’yi Vercel’de “sunucu” gibi kullanma iddiası yok; örnek **rewrite** ile istekleri Railway/Fly’daki API’ye yönlendirme |

### Örnek komutlar

```bash
# Hedefleri listele
thy-case-llm deploy list

# Bir hedefin açıklaması + hangi dosyaların yazılacağı
thy-case-llm deploy show railway
thy-case-llm deploy show vercel

# Önizleme (disk'e yazmaz, stdout'a basar)
thy-case-llm deploy init railway --dry-run

# Repo köküne yaz (mevcut dosya varsa hata verir)
thy-case-llm deploy init railway
thy-case-llm deploy init fly

# Üzerine yaz
thy-case-llm deploy init fly --force

# Başka dizine yaz (ör. monorepo alt paketi)
thy-case-llm deploy init railway --out ./backend

# Vercel rewrite hedefi (sonunda / olmasın)
thy-case-llm deploy init vercel --api-base-url https://api.örnek.com

# go.mod yoksa veya modül adını elle vermek için
thy-case-llm deploy init railway --module github.com/senin/projen
```

Yaygın flag’ler: `--dry-run`, `--force`, `--out <dir>`, `--port`, `--main-package`, `--health-path`, `--api-base-url` (vercel), `--module`. Tam liste: `thy-case-llm deploy` (alt komut yoksa yardım metni).
