<p align="center">
  <img src="./assets/turkiye-header.svg" alt="THY Case Study Backend" width="100%" />
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

Go tabanlı bir case-study backend projesi. Supabase Auth ve JWT claim tabanlı rol kontrolü kullanır; API yönlendirme/guard katmanı `gosupabase` ile çalışır.

## Built With gosupabase

Bu backend, benim geliştirdiğim `gosupabase` paketi üzerine kuruludur.
Paket; YAML-first endpoint yönetimi, Supabase JWT doğrulaması ve role-based route guard akışını sağlar.

- GitHub: [github.com/messivite/gosupabase](https://github.com/messivite/gosupabase)
- Go Package: [pkg.go.dev/github.com/messivite/gosupabase](https://pkg.go.dev/github.com/messivite/gosupabase)

## Mimari Özeti

- **Auth:** Supabase access token (`Authorization: Bearer <jwt>`)
- **Role modeli:** `public.user_roles` tablosu -> `custom_access_token_hook` -> JWT `claims.roles`
- **Rol kontrolü:** `api.yaml` içindeki `roles: [...]` alanları
- **Profil modeli:** `public.profiles` (`auth.users.id` ile 1:1)
- **Sohbet kalıcılığı:** Şu an bellek içi (in-memory) repository kullanılıyor (DB DSN gerektirmez)
- **Veritabanı:** Supabase Postgres (Auth tabloları + `public.profiles` + `public.user_roles` + RLS/policy)
- **LLM Provider Yönetimi:** `providers.yaml` + `.env` ile ayrıştırılmış konfig, `thy-case-llm` CLI ile yönetim

## DDD Katman Mimarisi (Faz 1)

```
┌─────────────────────────────────────────────────┐
│                  HTTP Katmanı                    │
│          internal/chat/handler.go                │
│      (DTO dönüşümleri, SSE akışı)               │
├─────────────────────────────────────────────────┤
│              Application Katmanı                 │
│       internal/application/chat/usecase.go       │
│  (iş kuralları, orkestrasyon, finalize akışı)    │
├─────────────────────────────────────────────────┤
│               Domain Katmanı                     │
│          internal/domain/chat/                   │
│  models · provider interface · repository i/f    │
│      errors · Role · StreamEvent · Request       │
├─────────────────────────────────────────────────┤
│           Infrastructure Katmanı                 │
│  internal/provider/  (OpenAI, Gemini adapter)    │
│  internal/repo/      (MemoryRepository)          │
│  internal/config/    (providers.yaml loader)     │
│  internal/auth/      (Supabase JWT adapter)      │
└─────────────────────────────────────────────────┘
```

**Yeni bir LLM provider eklemek için:**
1. `internal/provider/` altında yeni adapter dosyası oluştur (`domain.LLMProvider` interface'ini implemente et)
2. `providers.yaml`'a provider bilgisini ekle (veya `thy-case-llm provider add` kullan)
3. `cmd/api/main.go` içindeki `createProvider` switch'ine yeni case ekle

## Proje Yapısı

```
cmd/
  api/main.go              → API sunucu giriş noktası
  server/main.go           → gosupabase dev uyumlu giriş noktası
  thy-case-llm/main.go     → LLM provider yönetim CLI'ı
internal/
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
  config/                  → Konfig yükleme
    provider.go            → providers.yaml CRUD işlemleri
  chat/                    → HTTP handler katmanı
    handler.go             → REST + SSE endpoint'leri
  repo/                    → Repository implementasyonları
    memory_repository.go   → Bellek içi (in-memory) depo
  auth/                    → JWT doğrulama
    auth.go                → Middleware, context helpers
    supabase_adapter.go    → Supabase JWT/JWKS adapter
  app/
    server.go              → Router + middleware bağlama
providers.yaml             → LLM provider konfigürasyonu (non-secret)
api.yaml                   → Endpoint, auth ve rol kuralları
supabase/                  → Supabase CLI config, functions, migrations
```

## Provider Konfigürasyonu

LLM provider yönetimi iki katmanlı ayrışma prensibiyle çalışır:

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
    model: gemini-2.0-flash
    env_key: GEMINI_API_KEY
```

Sunucu başlarken `providers.yaml` okunur, her provider için ilgili env key kontrol edilir. Anahtarı eksik olan provider devre dışı bırakılır (uyarı loglanır).

## thy-case-llm CLI

LLM provider yönetimi için komut satırı aracı.

```bash
go run ./cmd/thy-case-llm help
```

Global komut olarak kullanmak için:

```bash
cd /Users/mustafaaksoy/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
```

### Komutlar

| Komut | Açıklama |
|-------|----------|
| `provider add` | Yeni provider ekle (interaktif veya `--name`, `--model`, `--env-key` flag'leri ile) |
| `provider list` | Kayıtlı provider'ları listele (varsayılan, model, env durumu) |
| `provider remove <name>` | Provider'ı kaldır |
| `provider set-default <name>` | Varsayılan provider'ı değiştir |
| `provider validate` | Tüm provider'ların env key kontrolünü yap |
| `doctor` | Provider + env + config için hızlı sağlık kontrolü |

### Kullanım Örnekleri

```bash
# Provider listele
thy-case-llm provider list

# Yeni provider ekle (interaktif)
thy-case-llm provider add

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

`zsh: command not found: thy-case-llm` hatası alırsanız:

1) Binary'yi kurun:

```bash
cd /Users/mustafaaksoy/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
```

2) Go bin path'i `PATH` içinde değilse ekleyin:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

3) Alternatif olarak kurulum gerektirmeden çalıştırın:

```bash
go run ./cmd/thy-case-llm doctor
```

## Environment

`.env.example`:

- `PORT` (varsayılan `8081`)
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`
- `SUPABASE_JWT_VALIDATION_MODE` (`auto` önerilir)
- `SUPABASE_ROLE_CLAIM_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`
- `PROVIDERS_CONFIG` (varsayılan `providers.yaml`, opsiyonel)

## Endpointler

| Metot | Endpoint | Auth | Açıklama |
|-------|----------|------|----------|
| `GET` | `/api/health` | Hayır | Sağlık kontrolü |
| `GET` | `/api/me` | Evet | JWT'deki kullanıcı bilgisi |
| `GET` | `/api/providers` | Evet | Aktif LLM provider'ları (default bilgisi dahil) |
| `POST` | `/api/chats` | Evet | Yeni sohbet oturumu oluştur |
| `GET` | `/api/chats` | Evet | Sohbet listesi |
| `GET` | `/api/chats/{chatID}` | Evet | Sohbet detayı + mesaj geçmişi |
| `POST` | `/api/chats/{chatID}/messages` | Evet | Mesaj gönder (non-stream) |
| `POST` | `/api/chats/{chatID}/stream` | Evet | Mesaj gönder (SSE stream) |

## Auth + Rol Akışı

1. Kullanıcı giriş yapar ve Supabase access token alır.
2. `custom_access_token_hook`, `public.user_roles` tablosunu okuyup token içindeki `claims.roles` alanını yazar.
3. API, `gosupabase` middleware ile JWT'yi doğrular.
4. Endpointte `roles: [...]` varsa token rolleri ile eşleşme kontrolü yapılır; eşleşmezse `403` döner.

Not: Rol değişikliğinden sonra yeni token alınması gerekir (logout/login veya token refresh).

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

Rol yönetimi akışı otomatik çalışır:

1. `public.user_roles` tablosuna `insert/update/delete` yapılır.
2. `trg_user_roles_sync_auth_metadata` trigger'ı tetiklenir.
3. Trigger, `auth.users.raw_app_meta_data.roles` alanını senkronlar.
4. `custom_access_token_hook`, yeni token üretiminde `public.user_roles` tablosunu okuyup `claims.roles` alanını yazar.

Not: Token immutable olduğu için mevcut token değişmez. Rol değişikliği sonrası kullanıcı yeni token almalıdır (logout/login veya refresh).

## Profiles Akışı

- `auth.users` kaydı oluştuğunda trigger ile `public.profiles` satırı açılır.
- `profiles.is_anonymous`, `auth.users.is_anonymous` alanından doldurulur/güncellenir.
- `display_name`, `avatar_url` gibi uygulama alanları uygulama katmanından update edilir.

## PostgreSQL (Supabase) Veri Modeli

- `auth.users`: Supabase Auth sistem tablosu (kayıt/giriş kimlik kaynağı)
- `public.profiles`: Uygulama profil verisi (`auth.users.id` ile birebir)
- `public.user_roles`: Kullanıcıya birden fazla rol atamak için ilişki tablosu
- RLS/policy kuralları Postgres seviyesinde uygulanır, API tarafı JWT claim ile yetki doğrular

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

Bu proje, PR kalite kontrolü için GitHub Actions + Codecov kullanır.

- CI adımları: `go mod tidy` kontrolü, `go build`, `go test`, `go vet`
- Coverage: test sonrası `coverage.out` üretilir ve Codecov'a yüklenir
- PR yönetimi: `codecov/patch` ve CI check'leri zorunlu kural olarak kullanılabilir
- Amaç: merge öncesi test geçişini ve kapsam düşüşlerini görünür kılmak

Codecov entegrasyonunun ilk doğrulama PR'ı: [trigger CI checkes #1](https://github.com/messivite/go-thy-case-study-backend/pull/1)

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

Chat oluştur:

```bash
curl -X POST "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"ilk chat"}'
```

Chat listele:

```bash
curl "http://localhost:8081/api/chats" \
  -H "Authorization: Bearer $TOKEN"
```

Chat detay:

```bash
curl "http://localhost:8081/api/chats/<CHAT_ID>" \
  -H "Authorization: Bearer $TOKEN"
```

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
    "model":"gemini-1.5-flash",
    "messages":[
      {"role":"user","content":"Bana kisa bir selamlama yaz"}
    ]
  }'
```

## Notlar

- `supabase/migrations` altında bazı migration dosyaları geçiş/legacy amacıyla tutuluyor.
- Kaynak rol modeli olarak `public.user_roles` kullanılmalıdır.
- `public.users` yeni geliştirmede kullanılmaz; profil için `public.profiles` kullanılır.
- **Faz 1:** DDD katman ayrımı yapıldı (domain → application → infrastructure). Provider yönetimi `providers.yaml` + `thy-case-llm` CLI ile standartlaştırıldı. Yeni provider eklemek sadece adapter dosyası + registry kaydı gerektirir.
