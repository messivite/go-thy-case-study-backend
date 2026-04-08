# Changelog

Önemli değişiklikler burada. Format [Keep a Changelog](https://keepachangelog.com/tr/1.0.0/) mantığında; sürümler [SemVer](https://semver.org/lang/tr/) ile uyumlu olmalı.

## [Unreleased]

### Added

- `CHANGELOG.md` ve `RELEASE_NOTES.md`.
- **`thy-case-llm deploy`:** `list`, `show <id>`, `init <id>` alt komutları; şablonlar `internal/deploy` içinde `go:embed` ile gömülü.
  - Hedefler: **`railway`** (`Dockerfile`, `railway.toml`), **`fly`** (aynı Dockerfile + `fly.toml`), **`vercel`** (örnek `vercel.json` + `deploy/VERCEL.md`; API’nin harici host’ta çalışması senaryosu).
  - `init` flag’leri: `--dry-run`, `--force`, `--out`, `--module`, `--port`, `--main-package`, `--health-path`, `--api-base-url` (vercel).
- `internal/deploy` birim testleri (şema yükleme, dry-run, dosya çakışması, `go.mod` modül tespiti).

### Changed

- `thy-case-llm` sürüm sabiti **v0.3.0**; `help` çıktısında komut sırası: `doctor` sonrasında `deploy` satırları.

### Documentation

- README: klon sonrası **`go install ./cmd/...`** ile `api`, `server`, `thy-case-llm` kurulumu; PATH notu.
- README: **`## Deploy`** bölümü (dosya sonu) — hedef tablosu, örnek komutlar, flag özeti; `thy-case-llm CLI` içinde deploy detayına iç link.

### Fixed

- Go modül yolu `github.com/messivite/go-thy-case-study-backend` olarak sabitlendi; `github.com/example/...` veya yanlış `messivite/thy-case-study-backend` import’ları CI’da kırılıyordu.

---

Aşağıdaki özet, anlamlı ilk semver tag’e kadar yapılan işler içindir. Yeni tag açtığında uygun sürüm başlığı altına taşıyabilirsin.

## Özet — Faz 1 & Faz 2 (tag öncesi)

### Added (Faz 1)

- Domain / application / infrastructure ayrımı (`internal/domain/chat`, `application/chat`, `provider`, `repo`, `chat` handler).
- `LLMProvider` + registry; OpenAI / Gemini adapter iskeleti → gerçek API (Faz 2).
- `providers.yaml` + `.env`; `thy-case-llm` CLI (`provider add|list|remove|set-default|validate`, `doctor`).
- Standart JSON hata cevapları (`internal/httpx`).

### Added (Faz 2)

- Provider template registry; `thy-case-llm templates list|show`, `provider add --template`.
- OpenAI / Gemini gerçek `Complete` + SSE `Stream`; token `usage` alanları.
- Supabase Postgres chat (`chat_sessions`, `chat_messages`); `CHAT_PERSISTENCE=supabase|memory`.
- `SupabaseRepository` (REST + service role).
- Usage normalizasyonu, structured log (`internal/observability`), HTTP retry / timeout, provider → HTTP hata eşlemesi.
- Provider hata durumunda user mesajının gereksiz persist edilmemesi (use case sırası).

### Tests

- Config, domain usage, httpx, provider registry/httpclient, repo memory testleri.
