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

1. Binary'yi kurun:

```bash
cd /Users/kullaniciAdiniz/Projects/go-thy-case-study-backend
go install ./cmd/thy-case-llm
```

2. Go bin path'i `PATH` içinde değilse ekleyin:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

Ya da direkt:

```bash
go run ./cmd/thy-case-llm doctor
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

### Observability (dosyaya log)

`internal/observability` paketinde yer alan `Info`, `Warn`, `Error` ile `LLM*` ile başlayan yardımcı fonksiyonlar, yapılandırılmış tek satırlık JSON günlük kaydı üretir (`ts`, `level`, `event`, `fields`, vb.). `OBSERVABILITY_LOG_FILE` ortam değişkeni tanımlandığında aynı kayıtlar belirtilen dosyaya eşzamanlı olarak eklenir (append). Süreç sonlanırken dosya tanıtıcısı `CloseFileLog` ile kapatılır; bu davranış `cmd/api` ve `cmd/server` giriş noktalarında uygulanmaktadır.

Örnek kayıt satırı:

```json
{"ts":"2026-04-09T12:00:00.123456789Z","level":"info","event":"llm.request","fields":{"provider":"openai","model":"gpt-4.1-mini","user_id":"...","session_id":"..."}}
```

Üretim ortamında dosya boyutunun yönetimi için logrotate, yan süreç (sidecar) veya merkezi günlük toplayıcı kullanılması önerilir. Bu uygulama içinde günlük döndürme (rotation) bilinçli olarak sağlanmamaktadır; mimari sade tutulmuştur.

### OpenTelemetry (trace)

**Kapsam:** Bu depoda dağıtılan izleme yapılandırması **temel (minimal)** düzeydedir. Gelen HTTP istekleri için sunucu tarafı span üretimi (`chi` yönlendirici ve `otelhttp` ile birlikte) ve **OTLP/HTTP** üzerinden dışa aktarım desteklenir. İlgili ortam değişkenleri tanımlı değilse izleme devreye girmez; bu sayede ek yüke maruz kalınmaz.

| Ortam değişkeni | Açıklama |
|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` veya `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Örnek değer: `http://localhost:4318` (Collector HTTP uç noktası). Tanımlı değilse dağıtılan trace özelliği etkin olmaz. |
| `OTEL_SERVICE_NAME` | Varsayılan: `thy-case-study-api`. |

Örnek Collector yapılandırması bu depoda **`otel/collector.yaml`** dosyasında bulunmaktadır. `otelcol` ikilisi hangi çalışma dizininden çalıştırılırsa çalıştırılsın, yapılandırma dosyasına **mutlak yol** verilmesi gerekir:

```bash
./otelcol --config=/ABSOLUTE/PATH/thy-case-study-backend/otel/collector.yaml
```

`config.yaml: no such file` iletisinin görülmesi, komutun geçerli çalışma dizininde yapılandırma aranmasından kaynaklanır; `--config=` bağımsız değişkeni ile depo içindeki dosyanın tam yolu belirtilmelidir.

#### OpenTelemetry Collector kurulumu (yerel doğrulama)

1. **Dağıtım** — Güncel sürümler [OpenTelemetry Collector releases](https://github.com/open-telemetry/opentelemetry-collector-releases/releases) sayfasından edinilebilir. Bu depodaki örnek yapılandırma, **Core** dağıtımı (`otelcol`) ile uyumludur; macOS için mimariye uygun `otelcol_*_darwin_arm64.tar.gz` veya `*_amd64` paketi seçilebilir. Ek alıcı veya dışa aktarıcı gereksinimi bulunması hâlinde aynı kaynak üzerinden **[otelcol-contrib](https://github.com/open-telemetry/opentelemetry-collector-releases/releases)** dağıtımı değerlendirilebilir. Kurulumun özeti için bkz. [Collector installation](https://opentelemetry.io/docs/collector/installation/).
2. **Kurulum** — Arşiv açıldıktan sonra `otelcol` ikilisi uygun bir dizine yerleştirilir. macOS üzerinde ilk çalıştırmada güvenlik uyarısı çıkması hâlinde Finder üzerinden sağ tıklayıp **Açı** seçeneği kullanılabilir veya `xattr -dr com.apple.quarantine ./otelcol` komutu uygulanabilir.
3. **Çalıştırma** — Collector, ayrı bir terminal oturumunda sürekli çalışacak şekilde başlatılır:

   ```bash
   ./otelcol --config=/ABSOLUTE/PATH/thy-case-study-backend/otel/collector.yaml
   ```

   Günlük çıktısında `Starting HTTP server` ile `endpoint: "[::]:4318"` ifadelerinin yer alması beklenir.
4. **API tarafı** — `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318` ve gerekmesi hâlinde `OTEL_SERVICE_NAME` değerleri tanımlanmalıdır. Bu değerler depo kökündeki `.env` dosyasına yazılabilir (`cmd/api` başlangıçta yükler) veya kabuk ortamı / IDE ortam değişkenleri üzerinden iletilebilir.
5. **Doğrulama** — `GET /health` ve `GET /api/health` uç noktaları trace üretmez. `GET /api/providers` veya `GET /api/chats` gibi kimlik doğrulaması gerektiren bir uç noktaya geçerli bir JWT ile istek gönderildiğinde, trace çıktısı Collector sürecinin standart çıktısında (`debug` exporter) gözlemlenebilir.

**Örnek ekran görüntüsü** — Bu projeden `GET /api/chats` isteğine karşılık Collector `debug` exporter çıktısı:

![OpenTelemetry Collector debug exporter ile gelen trace örneği](https://i.ibb.co/VWrfvghD/Screenshot-2026-04-09-at-02-22-32.png)

*(Doğrudan bağlantı: [Screenshot-2026-04-09-at-02-22-32.png](https://i.ibb.co/VWrfvghD/Screenshot-2026-04-09-at-02-22-32.png))*

Üretim ortamında Collector yapılandırmasına Jaeger, Grafana Tempo veya satıcı tarafı OTLP uç noktaları gibi ek **exporter** tanımları eklenmesi uygun olur; yerel doğrulama için yalnızca `debug` exporter ile konsol çıktısının incelenmesi genellikle yeterlidir.

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

1. `user_roles` tablosuna `insert/update/delete` işlemi yapılır.
2. `trg_user_roles_sync_auth_metadata` trigger'ı çalışır.
3. `auth.users.raw_app_meta_data.roles` alanı güncellenir.
4. Kullanıcı yeni token aldığında hook yeniden `user_roles` tablosunu okuyup JWT içindeki `claims.roles` alanını yazar.

Eski token'ın içeriği sonradan değişmez; rol değişikliği sonrası mutlaka yeni token alınmalıdır.

## Profiles Akışı

- User oluştuğunda trigger ile `public.profiles` satırı otomatik açılır.
- `is_anonymous` değeri `auth` tarafındaki flag'den taşınır.
- `display_name`, `avatar_url` gibi profil alanları uygulama tarafından güncellenir.

## PostgreSQL (Supabase) Veri Modeli

- `auth.users` -> Kimlik kayıtları
- `public.profiles` -> Profil tablosu (`auth.users` ile 1:1)
- `public.user_roles` -> Kullanıcıya bağlı çoklu rol ilişkisi
- `public.chat_sessions`, `public.chat_messages` -> Sohbet verisi (migration dosyalarında tanımlı)
- `public.llm_interaction_log` -> LLM audit ve usage logu
- `public.llm_quota_defaults`, `public.user_llm_usage_quota` -> Kota konfigürasyonu
- RLS kuralları Postgres/Supabase tarafında uygulanır; API tarafında JWT doğrulaması ve role enforcement devam eder.

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
