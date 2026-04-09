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

| Ad | Açıklama | Link |
|---|---|---|
| PROD Base URL | Canlı ortam ana adresi (apisiz) | [http://go-thy-case-study-backend-production.up.railway.app/](http://go-thy-case-study-backend-production.up.railway.app/) |
| PROD API URL | Canlı ortam API adresi (`/api`) | [http://go-thy-case-study-backend-production.up.railway.app/api](http://go-thy-case-study-backend-production.up.railway.app/api) |
| PROD Swagger UI | Canlı ortamda API dokümantasyonu ve endpoint deneme ekranı | [http://go-thy-case-study-backend-production.up.railway.app/docs-thy-case-study-backend](http://go-thy-case-study-backend-production.up.railway.app/docs-thy-case-study-backend) |
| DEV Base URL | Lokal geliştirme ortamı ana adresi (apisiz) | [http://localhost:8081/](http://localhost:8081/) |
| DEV API URL | Lokal geliştirme API adresi (`/api`) | [http://localhost:8081/api](http://localhost:8081/api) |
| DEV Swagger UI | Lokal ortamda API dokümantasyonu ve endpoint test ekranı | [http://localhost:8081/docs-thy-case-study-backend](http://localhost:8081/docs-thy-case-study-backend) |

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
- `SWAGGER_PUBLIC_PATH` (varsayilan `/docs-thy-case-study-backend`)

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

## API Dokümantasyonu (Swagger UI)

Uygulama çalışırken OpenAPI 3.1 tabanlı interaktif API dokümantasyonuna erişebilirsiniz:

```
http://localhost:8081/docs-thy-case-study-backend
```

Bu endpoint auth gerektirmez. Path, `SWAGGER_PUBLIC_PATH` env değişkeniyle değiştirilebilir.

| Path | Format |
|---|---|
| `{SWAGGER_PUBLIC_PATH}/` | Swagger UI (interaktif) |
| `{SWAGGER_PUBLIC_PATH}/openapi.json` | OpenAPI 3.1 JSON |
| `{SWAGGER_PUBLIC_PATH}/openapi.yaml` | OpenAPI 3.1 YAML |

### Swagger/OpenAPI Otomatik Üretim

Swagger dokümanlarının kaynağı `api.yaml` dosyasıdır. Endpoint tanımlarını burada güncellersin; çıktı dosyaları otomatik üretilir.

Komut:

```bash
make openapi
```

Bu komut:
- `scripts/generate_openapi.go` scriptini çalıştırır.
- `docs/openapi.yaml` ve `docs/openapi.json` dosyalarını yeniden üretir.

Önerilen akış:
1. `api.yaml` içinde endpoint/handler/auth tanımını güncelle.
2. İlgili handler kodunu güncelle.
3. `make openapi` çalıştır.
4. Üretilen `docs/openapi.yaml` ve `docs/openapi.json` dosyalarını commit et.

### CI Validasyonu (Drift Kontrolü)

CI sürecinde OpenAPI dosyalarının güncel kalması doğrulanır:
- `go run ./scripts/generate_openapi.go`
- `git diff --exit-code docs/openapi.yaml docs/openapi.json`

Eğer `api.yaml` değişmiş ama üretilen dosyalar commit edilmemişse pipeline hata verir.

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
