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

**Go modülü:** `github.com/messivite/go-thy-case-study-backend` — `go.mod` ile bütün `import` satırları bununla uyumlu; `github.com/example/...` kullanılmıyor (CI’da yanlış modül aranıp patlamasın diye).

**Sürüm notları:** [CHANGELOG.md](CHANGELOG.md) · [RELEASE_NOTES.md](RELEASE_NOTES.md) (tag / GitHub Release akışı)

## Built With gosupabase

`gosupabase` paketini ben yazdım; bu repo onun üstünde. YAML’dan endpoint, JWT doğrulama, role guard işi oradan geliyor.

- GitHub: [github.com/messivite/gosupabase](https://github.com/messivite/gosupabase)
- Go Package: [pkg.go.dev/github.com/messivite/gosupabase](https://pkg.go.dev/github.com/messivite/gosupabase)

## Mimari Özeti

- **Auth:** Supabase access token (`Authorization: Bearer <jwt>`)
- **Roller:** `public.user_roles` → hook → JWT’de `claims.roles`
- **Route rolleri:** `api.yaml` içindeki `roles: [...]`
- **Profil:** `public.profiles`, `auth.users` ile 1:1
- **Chat:** Varsayılan Supabase Postgres (REST + service role); istersen `CHAT_PERSISTENCE=memory` ile sadece RAM
- **DB:** Supabase Postgres (auth + `profiles` + `user_roles` + chat tabloları, RLS)
- **LLM:** `providers.yaml` + `.env`, yönetim için `thy-case-llm`

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
api.yaml                   → Endpoint, auth ve rol kuralları
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
cd /Users/kullaniciAdiniz/Projects/thy-case-study-backend
go install ./cmd/thy-case-llm
```

### Önemli not (güncel binary)

`templates list`, `templates show`, `provider add --template …` Faz 2’den beri var. Eski `go install` binary’si kalırsa şunu görürsün:

```text
unknown command: templates
```

`thy-case-llm help` içinde de `templates` yoksa PATH’teki binary eski demektir.

**Ne yapayım:** Projede tekrar `go install`, `$(go env GOPATH)/bin` PATH’te olsun:

```bash
cd /Users/kullaniciAdiniz/Projects/thy-case-study-backend
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
cd /Users/kullaniciAdiniz/Projects/thy-case-study-backend
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
- `CHAT_PERSISTENCE` — yazmazsan `supabase`; `memory` dersen RAM. `SUPABASE_URL` / `SUPABASE_SERVICE_ROLE_KEY` eksikse yine memory’ye düşüyor (`cmd/api/main.go`).
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
| `POST` | `/api/chats` | Evet | Yeni sohbet; gövdede isteğe bağlı `provider`, `model` (`providers.yaml` / `GET /api/providers`). Cevap: `id`, `provider`, `model` (session default’ları) |
| `GET` | `/api/chats` | Evet | Sohbet listesi |
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
      {"role":"user","content":"Bana kisa bir selamlama yaz"}
    ]
  }'
```

## Notlar

- `supabase/migrations` içinde bazı dosyalar eski geçişlerden kaldı, silmiyorum referans için.
- Rol kaynağı: `public.user_roles`.
- Yeni iş için `public.users` değil, `public.profiles`.
- **Faz 1:** Domain / application / infra ayrımı, `providers.yaml` + `thy-case-llm`, yeni LLM için adapter + registry.
- **Faz 2:** Template CLI (`templates list/show`), gerçek OpenAI/Gemini çağrıları, chat’i Postgres’e yazma, usage normalize, log tarafı, retry/timeout, provider hatalarını HTTP’ye map etme.

### Faz 3 — sonra bakacağımız işler

#### Self-hosted / özel endpoint

Kendi makinemde veya şirket gateway’inde model çalıştırırsam şu an kod sadece OpenAI’nin sabit URL’ine gidiyor. Faz 3’te base URL’yi env veya `providers.yaml`’dan verebilir hale getirmek istiyorum (vLLM, LiteLLM proxy vs.). Gerekirse ek header. CLI’da da bu endpoint’i tanımlama. “Sadece farklı model id” ile “farklı host” ayrımını dokümanda net yazarım.

#### Token kotası — günlük / haftalık, global + kullanıcıya özel

OpenAI’nin döndürdüğü `usage` token’larını kullanıp ürün içi limit koymak:

- Proje default’u: günlük + haftalık limit tek yerde (tablo veya settings kaydı); yeni üye için başlangıç değeri buradan.
- Kullanıcıya özel: ayrı tablo, `user_id` FK; günlük/haftalık override.
- Register’da default’u kullanıcı kotasına yazan trigger veya hook.
- İstekte (veya LLM cevabı geldikten sonra) sayaç güncelle; limit dolunca anlamlı `429` + `code`.
- Admin panelden user bulup kotayı düzenleme; satır yoksa default’tan üretme kuralı — detayını sonra netleştiririz.
