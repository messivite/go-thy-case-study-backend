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

## Proje Yapısı

- `cmd/api/main.go` - API başlangıç noktası
- `cmd/server/main.go` - `gosupabase dev` ile uyumlu başlangıç noktası
- `internal/app` - HTTP sunucu ve route bağlama katmanı
- `internal/auth` - Supabase JWT doğrulama adaptörü
- `internal/chat` - Chat handler/service katmanı
- `internal/repo` - Bellek içi repository implementasyonu
- `supabase/` - Supabase CLI config, functions ve migrations
- `api.yaml` - Endpoint, auth ve rol kuralları

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

## Endpointler

`api.yaml` altında:

- `GET /api/health` (auth yok)
- `GET /api/me` (auth var)
- `GET /api/providers` (auth var)
- `GET /api/sessions` (auth var)
- `POST /api/sessions` (auth var)
- `GET /api/sessions/{sessionID}/messages` (auth var)
- `POST /api/sessions/{sessionID}/messages` (auth var)

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

## Yerelde Çalıştırma

```bash
gosupabase dev
```

veya:

```bash
go run ./cmd/api
```

## Notlar

- `supabase/migrations` altında bazı migration dosyaları geçiş/legacy amacıyla tutuluyor.
- Kaynak rol modeli olarak `public.user_roles` kullanılmalıdır.
- `public.users` yeni geliştirmede kullanılmaz; profil için `public.profiles` kullanılır.
