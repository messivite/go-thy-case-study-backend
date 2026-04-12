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
  <a href="https://hub.docker.com/r/messivite47/thy-case-study-backend">
    <img src="https://img.shields.io/badge/Docker%20Hub-messivite47%2Fthy--case--study--backend-2496ED?logo=docker&logoColor=white&style=for-the-badge" alt="Docker Hub: messivite47/thy-case-study-backend" />
  </a>
</p>

# THY için Case Study Kapsamında Hazırlanan Backend Side Go Projesi

| Ad | Açıklama | Link |
|---|---|---|
| PROD Base URL | Canlı ortam kök adresi (`/api` öneki yok) | [http://go-thy-case-study-backend-production.up.railway.app/](http://go-thy-case-study-backend-production.up.railway.app/) |
| PROD API URL | Canlı ortam API adresi (`/api`) | [http://go-thy-case-study-backend-production.up.railway.app/api](http://go-thy-case-study-backend-production.up.railway.app/api) |
| PROD Swagger UI | Canlı ortamda API dokümantasyonu ve endpoint deneme ekranı | [http://go-thy-case-study-backend-production.up.railway.app/docs-thy-case-study-backend](http://go-thy-case-study-backend-production.up.railway.app/docs-thy-case-study-backend) |
| DEV Base URL | Lokal geliştirme kök adresi (`/api` öneki yok) | [http://localhost:8082/](http://localhost:8082/) |
| DEV API URL | Lokal geliştirme API adresi (`/api`) | [http://localhost:8082/api](http://localhost:8082/api) |
| DEV Swagger UI | Lokal ortamda API dokümantasyonu ve endpoint test ekranı | [http://localhost:8082/docs-thy-case-study-backend](http://localhost:8082/docs-thy-case-study-backend) |

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
- **HTTP yanıt önbelleği (opsiyonel):** `GET /api/chats` ve `GET /api/chats/{id}/messages` için env ile açılır (`memory` veya `redis`)
- **LLM:** `providers.yaml` + environment variable anahtarları
- **Desteklenen modeller:** Açılışta kod tabanlı katalog ile veritabanı (veya bellek modunda in-process liste) senkron; istemci seçenekleri bu kaynaktan gelir, katalogda aktif olmayan modele yapılan çağrılar reddedilir
- **Kota ve audit:** Supabase tabloları + trigger + RPC

## Proje yapısı

```text
cmd/
  api/main.go              -> Üretim tipi API sunucusunun giriş noktası
  server/main.go           -> Yerel geliştirme (gosupabase dev) giriş noktası
  thy-case-llm/main.go     -> LLM sağlayıcı ve dağıtım şablonları için CLI aracı
internal/
  application/chat/        -> Use case katmanı
  app/                     -> HTTP yönlendirme, Swagger UI, landing
  auth/                    -> JWT doğrulama ve istek bağlamındaki kullanıcı bilgisi
  cache/                   -> HTTP yanıt önbelleği (bellek veya Redis)
  chat/                    -> Sohbet ve ilgili HTTP handler’lar
  config/                  -> Sağlayıcı yapılandırması ve şablonlar
  catalog/                 -> Kayıtlı sağlayıcılar + yerleşik şablonlardan “seçilebilir model” kataloğu üretimi
  deploy/                  -> Dağıtım hedefleri (list / show / init) ve şablonlar
  domain/chat/             -> Domain modelleri, repository arayüzleri, hatalar
  dotenv/                  -> Yerel .env yükleme yardımcıları
  httpx/                   -> HTTP hata yanıtları
  observability/           -> Yapılandırılmış log ve OpenTelemetry izleme
  provider/                -> OpenAI / Gemini sağlayıcıları ve kayıt defteri (registry)
  repo/                    -> Supabase ve bellek içi repository uygulamaları
  swagger/                 -> OpenAPI belgelerinin sunumu
  landing/                 -> Kök path karşılama sayfası
providers.yaml             -> LLM sağlayıcı tanımları (gizli olmayan ayarlar)
api.yaml                   -> Uç nokta tanımları ve rol kuralları
supabase/                  -> Veritabanı migration’ları, edge function’lar ve Supabase yapılandırması
```

## Provider konfigürasyonu

Sağlayıcı meta verileri (isim, model, env anahtarı referansı) ile gizli anahtarlar ayrı tutulur:

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

## Desteklenen LLM modelleri (katalog ve yönetim)

Hangi modellerin **seçilebilir** ve **kullanıma açık** sayılacağı, tek bir “yaşayan katalog” üzerinden yönetilir; böylece arayüzde gösterilen liste ile sunucunun kabul ettiği modeller uyumlu kalır.

**Kaynak (gerçek):** Çalışma anında registry’de **etkin** olan sağlayıcılar ile bu projedeki **yerleşik şablonlar** birleştirilir. Şablonda tanımlı model listesi varsa o liste kullanılır; özel veya şablonsuz bir sağlayıcı için en azından yapılandırmadaki varsayılan model satırı kataloğa eklenir. Yani katalog, “hangi API anahtarları yüklü” ve “kodda hangi modeller tanımlı” gerçeğine bağlıdır; `providers.yaml`’daki satırlar tek başına yeterli değildir, env ile devre dışı kalan sağlayıcı katalogda da yer almaz.

**Senkronizasyon:** API süreci her başladığında bu katalog, kalıcı modda Supabase’teki ilgili tabloya yazılır (güvenli tanımlı bir RPC ile toplu güncelleme). Tabloda artık yeni listede bulunmayan sağlayıcı/model çiftleri **pasif** işaretlenir; böylece bir modeli koddan veya şablondan kaldırdığınızda bir sonraki deploy veya restart sonrası hem liste hem doğrulama tarafı güncellenir. Aynı model sonra tekrar katalog üretimine girerse bir sonraki senkronla yeniden aktifleşir. Tamamen kalıcı olarak kapatmak istediğiniz senaryoda modeli üretim listesinden çıkarmanız gerekir; operasyon ekibi doğrudan veritabanında pasif bayrağı kullanarak da müdahale edebilir, ancak bir sonraki uygulama senkronu katalogda hâlâ varsa satırı tekrar açabilir.

**İstemci deneyimi:** Kimliği doğrulanmış kullanıcılar, arayüzde gösterecekleri model listesini bu katalogdan (API üzerinden) alır; provider özetinden ayrı tutulur çünkü biri “hangi entegrasyonlar açık”, diğeri “hangi model kimliği seçilebilir” sorusunu yanıtlar.

**İstek doğrulaması:** Sohbet açma, mesaj gönderme, stream ve toplu sync gibi LLM çağrısı yapan tüm akışlarda, çözümlenen sağlayıcı ve kullanılacak efektif model katalogda **aktif** değilse istek anlamlı bir hata ile reddedilir. Böylece kullanıcı eski bir istemci sürümünden kaldırılmış veya devre dışı bir model kimliği gönderse bile sunucu tutarlı şekilde “bu model artık kullanılamıyor” mesajına yakın bir davranış sergiler.

**Bellek modu:** `CHAT_PERSISTENCE` bellek olarak seçildiğinde aynı senkron ve doğrulama mantığı yalnızca süreç içi bellekte çalışır; Postgres’e yazılmaz. Yerel geliştirmede migration uygulamadan da bu akış test edilebilir.

**RLS:** Kalıcı ortamda tabloya yalnızca servis rolü tam yazım yapar; oturum açmış son kullanıcılar için satır düzeyi güvenlik, yalnızca **aktif** kayıtların okunmasına izin verecek şekilde tanımlıdır (doğrudan Supabase istemcisi ile okuma yapan uygulamalar için).

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
| `cache config` | HTTP yanıt önbelleği için `.env` içine `CACHE_*` / `REDIS_*` yazar (interaktif) |
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

# Claude (Anthropic API — ANTHROPIC_API_KEY; providers.yaml'da name: claude)
thy-case-llm provider add --template claude --set-default
# veya kayıt adı anthropic olsun istersen:
thy-case-llm provider add --template anthropic --set-default

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

- `PORT` (varsayılan `8082`; `cmd/server` ve Docker örnekleri ile uyumlu)
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
- Önbellek: `CACHE_ENABLED`, `CACHE_BACKEND`, `CACHE_TTL_*`, `REDIS_*` (detay için aşağıdaki **Cache** bölümüne bak)

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

## Cache

Bunu ben **sunucu tarafında**, sık tekrarlanan okuma isteklerini hafifletmek için ekledim: aynı kullanıcı aynı endpoint’e tekrar geldiğinde (TTL süresi içinde) cevabı doğrudan bellekten veya Redis’ten dönüyoruz; böylece her seferinde DB’ye gitmek zorunda kalmıyorsun. Bu bir **HTTP yanıt önbelleği**; istemci tarafında ekstra bir şey yapmana gerek yok.

**Ne önbelleniyor?**

- `GET /api/chats` — sohbet listesi (legacy dizi veya sayfalı cevap; **query string** farklıysa ayrı anahtar, birbirinin üstüne binmez).
- `GET /api/chats/{chatID}/messages` — sayfalı mesaj listesi (yine query’ye göre ayrı entry).

**Nasıl yönetiyorum?**

- Tamamen **ortam değişkeni** ile. `api.yaml` veya Swagger’da bir “cache toggle” yok; deploy ortamında env set ediyorsun, sunucu açılışında `internal/cache.FromEnv()` okuyor.
- İstersen elle `.env` yazarsın, istersen repodaki CLI ile interaktif doldurursun:

```bash
thy-case-llm cache config
```

Bu komut `.env` içine (veya `ENV_FILE` neyse ona) işaretli bir blok halinde `CACHE_*` / gerekiyorsa `REDIS_*` yazar; tekrar çalıştırınca aynı bloğu günceller.

**Backend seçenekleri**

| Değer | Ne zaman? |
| --- | --- |
| `memory` (varsayılan) | Tek process / tek makine; en basit kurulum, ek servis yok. |
| `redis` | Birden fazla API instance veya Redis’i merkezi cache olarak kullanmak istediğinde; tüm pod’lar aynı Redis’i görür. |

**Invalidation (ne zaman sıfırlanıyor?)**

Veri değişince ilgili kullanıcı için liste ve/veya o sohbetin mesaj önbelleği temizleniyor: yeni sohbet, mesaj gönderme, stream’in bitmesi, sync, soft-delete vb. Böylece eski JSON’u yanlışlıkla uzun süre servis etmiyoruz.

**Örnek env (memory, açık)**

```bash
CACHE_ENABLED=true
CACHE_BACKEND=memory
CACHE_TTL_CHAT_LIST_SEC=20
CACHE_TTL_CHAT_MESSAGES_SEC=15
```

**Örnek env (redis)**

```bash
CACHE_ENABLED=true
CACHE_BACKEND=redis
CACHE_TTL_CHAT_LIST_SEC=20
CACHE_TTL_CHAT_MESSAGES_SEC=15
REDIS_ADDR=127.0.0.1:6379
REDIS_DB=0
# REDIS_PASSWORD=   # gerekiyorsa
```

Kapalı tutmak için `CACHE_ENABLED`’ı verme veya `false` yapman yeterli; geri kanıt akışı eskisi gibi çalışır.

## API Dokümantasyonu (Swagger UI)

Uygulama çalışırken OpenAPI 3.1 tabanlı interaktif API dokümantasyonuna erişebilirsiniz:

```
http://localhost:8082/docs-thy-case-study-backend
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
| `GET` | `/api/me/usage` | Evet | Günlük/haftalık token kullanım-kota özeti |
| `GET` | `/api/chats/search?q=...&limit=...&cursor=...` | Evet | Kullanıcıya ait sohbetlerde title + user/assistant mesaj araması (cursor pagination) |
| `GET` | `/api/providers` | Evet | Aktif provider listesi ve default bilgi |
| `POST` | `/api/chats` | Evet | Yeni sohbet; isteğe bağlı `content` ile ilk mesajı aynı istekte gönderip `assistantMessage` alabilirsin |
| `GET` | `/api/chats` | Evet | Sohbet listesi (opsiyonel `limit` + `cursor` ile infinite-scroll pagination) |
| `GET` | `/api/chats/{chatID}` | Evet | Sohbet ve mesaj detaylarını döner |
| `DELETE` | `/api/chats/{chatID}` | Evet | Sohbeti soft-delete eder (`deleted_at` set edilir) |
| `GET` | `/api/chats/{chatID}/messages` | Evet | Mesajları `direction=older|newer` + `cursor` ile iki yönlü sayfalar |
| `DELETE` | `/api/chats/{chatID}/messages/{messageID}` | Evet | Kullanıcının kendi `user` mesajını soft-delete eder |
| `POST` | `/api/chats/{chatID}/messages` | Evet | Non-stream mesaj gönderir |
| `POST` | `/api/chats/{chatID}/stream` | Evet | SSE stream mesaj gönderir |

### Soft Delete davranışı (`DELETE /api/chats/{chatID}`)

- Fiziksel silme yapılmaz; `public.chat_sessions.deleted_at` alanı UTC timestamp ile işaretlenir.
- Mesaj soft-delete için `public.chat_messages.deleted_at` kullanılır.
- Soft-delete edilmiş oturumlar listeleme, arama ve mesaj sayfalama sonuçlarına dahil edilmez.
- Soft-delete edilmiş mesajlar da chat detay ve mesaj listesi sonuçlarında görünmez.
- Yetki kontrolü zorunludur: sadece JWT'deki kullanıcıya ait `chatID` için işlem yapılır.
- Aynı oturuma tekrar delete çağrısı idempotent olarak `404` dönebilir (aktif kayıt bulunamaz).

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

Yerel API’yi (varsayılan port **8082**) internet üzerinden paylaşmak veya mobil / başka bir makineden test etmek için [ngrok](https://ngrok.com/) kullanabilirsiniz:

```bash
ngrok http 8082
```

Komut, `localhost:8082`’ye tünel açan geçici bir HTTPS URL üretir; bu adresi istemci veya webhook tarafında taban URL olarak kullanın (`/api` yolu aynı kalır).

## Deploy

CLI v0.3.0+ ile deploy şablonları üretilir:

```bash
thy-case-llm deploy list
thy-case-llm deploy show railway
thy-case-llm deploy show docker
thy-case-llm deploy init railway --dry-run
thy-case-llm deploy init railway
thy-case-llm deploy init docker
```

Desteklenen hedefler:

| id | Üretilen dosyalar | Not |
|---|---|---|
| `railway` | `Dockerfile`, `railway.toml` | Varsayılan health path `/health` |
| `fly` | `Dockerfile`, `fly.toml` | Benzer Docker tabanlı kurulum |
| `docker` | `Dockerfile`, `docker-compose.yml`, `.dockerignore` | Yerel Docker/Compose çalıştırma ve registry image hazırlığı |
| `vercel` | `vercel.json`, `deploy/VERCEL.md` | Rewrite tabanlı yönlendirme senaryosu |

Yaygin flag'ler: `--dry-run`, `--force`, `--out`, `--port`, `--main-package`, `--health-path`, `--api-base-url`, `--module`.

### Docker ile Çalıştırma

Bu repo Docker ile **yerelde** doğrudan ayağa kalkabilir; bunun için image'ı bir yere push etmen gerekmez.

```bash
docker compose up --build
```

API:

```text
http://localhost:8082/api
```

Sadece image üretmek için:

```bash
docker build -t thy-case-study-backend:local .
docker run --rm -p 8082:8082 --env-file .env thy-case-study-backend:local
```

### Docker Hub (public image)

İmaj sayfası: [messivite47/thy-case-study-backend](https://hub.docker.com/r/messivite47/thy-case-study-backend)

Hub’dan çekmek:

```bash
docker pull messivite47/thy-case-study-backend:latest
```

Çalıştırmak (örnek; `.env` veya ortam değişkenlerini kendi ortamına göre ver):

```bash
docker run --rm -p 8082:8082 --env-file .env messivite47/thy-case-study-backend:latest
```

İmajı güncelleyip Hub’a göndermek (`tagname` yerine sürüm veya `latest`):

```bash
docker login
docker tag thy-case-study-backend:local messivite47/thy-case-study-backend:tagname
docker push messivite47/thy-case-study-backend:tagname
```

Uzak bir platforma (ECS/Render/Railway/Fly) kendi registry tag’in ile çıkmak istersen aynı kalıbı kullan: `docker build -t <registry>/<image>:<tag> .` ve `docker push <registry>/<image>:<tag>`.

### Deploy: `docker` hedefi ne işe yarar?

`thy-case-llm deploy init docker` komutu, repoya **üretime uygun Docker dosyalarını** yazar veya günceller (şablonlar `internal/deploy/bundle` altında). Temel fikir:

| Dosya | Rolü |
| --- | --- |
| `Dockerfile` | Go binary’yi çok aşamalı build edip küçük bir Linux imajında çalıştırır. |
| `docker-compose.yml` | Yerelde tek komutla (`docker compose up`) API’yi ayağa kaldırır; port ve `PORT` env ile uyumludur. |
| `.dockerignore` | Gereksiz dosyaların imaja girmesini azaltır; build daha hızlı ve imaj daha küçük olur. |

**Yerel:** `docker compose up --build` — push gerekmez, `http://localhost:8082` üzerinden erişirsin.  
**Uzak:** Aynı `Dockerfile` ile `docker build` + registry’ye `docker push`; platform (Railway, Fly, ECS, vb.) bu image’i çalıştırır.

<p align="center">
  <a href="https://hub.docker.com/r/messivite47/thy-case-study-backend" title="Docker Hub: messivite47/thy-case-study-backend">
    <img src="https://cdn.simpleicons.org/docker/2496ED" alt="Docker Hub" width="56" height="56" />
  </a>
</p>
