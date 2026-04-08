<p align="center">
  <img src="./assets/turkiye-header.svg" alt="THY Case Study Backend" width="100%" />
</p>
<p align="center">
  <a href="https://pkg.go.dev/github.com/messivite/gosupabase">
    <img src="https://pkg.go.dev/badge/github.com/messivite/gosupabase.svg" alt="Go Reference: gosupabase" />
  </a>
  <a href="https://github.com/messivite/gosupabase/releases">
    <img src="https://img.shields.io/github/v/release/messivite/gosupabase?label=gosupabase%20release" alt="gosupabase release" />
  </a>
  <a href="https://github.com/messivite/gosupabase/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/messivite/gosupabase" alt="License" />
  </a>
</p>

# Thy Case Study Backend

Go tabanlı bir case-study backend projesi. Supabase Auth ve JWT claim tabanlı rol kontrolü kullanır; API yönlendirme/guard katmanı `gosupabase` ile çalışır.

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
