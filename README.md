<p align="center">
  <img src="./assets/turkiye-header.svg" alt="THY Case Study Backend" width="100%" />
</p>

# Thy Case Study Backend

Go tabanli case-study backend. Supabase Auth + JWT claim tabanli role kontrolu kullanir, API routing/guard katmani `gosupabase` ile calisir.

## Mimari Ozeti

- **Auth:** Supabase access token (`Authorization: Bearer <jwt>`)
- **Role modeli:** `public.user_roles` tablosu -> `custom_access_token_hook` -> JWT `claims.roles`
- **Role enforcement:** `api.yaml` icindeki `roles: [...]` alanlari
- **App profile modeli:** `public.profiles` (`auth.users.id` ile 1:1)
- **Chat persistence:** su an in-memory repository (DB DSN gerektirmez)
- **Veritabani:** Supabase Postgres (Auth tablolari + `public.profiles` + `public.user_roles` + RLS/policy)

## Proje Yapisi

- `cmd/api/main.go` - API bootstrap
- `cmd/server/main.go` - `gosupabase dev` uyumlu bootstrap girisi
- `internal/app` - HTTP server ve route wiring
- `internal/auth` - Supabase JWT dogrulama adaptoru
- `internal/chat` - chat handler/service
- `internal/repo` - in-memory repository implementasyonu
- `supabase/` - Supabase CLI config, functions, migrations
- `api.yaml` - endpoint, auth ve rol kurallari

## Environment

`.env.example`:

- `PORT` (varsayilan `8081`)
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`
- `SUPABASE_JWT_VALIDATION_MODE` (`auto` onerilir)
- `SUPABASE_ROLE_CLAIM_KEY`
- `OPENAI_API_KEY`
- `GEMINI_API_KEY`

## Endpointler

`api.yaml` altinda:

- `GET /api/health` (auth yok)
- `GET /api/me` (auth var)
- `GET /api/providers` (auth var)
- `GET /api/sessions` (auth var)
- `POST /api/sessions` (auth var)
- `GET /api/sessions/{sessionID}/messages` (auth var)
- `POST /api/sessions/{sessionID}/messages` (auth var)

## Auth + Role Akisi

1. Kullanici login olur, Supabase access token alir.
2. `custom_access_token_hook`, `public.user_roles` tablosunu okuyup token `claims.roles` alanini yazar.
3. API `gosupabase` middleware ile JWT'yi dogrular.
4. Endpointte `roles: [...]` varsa token rol(ler)i ile eslestirir, uyusmazsa `403`.

Not: Rol degisikliginden sonra yeni token alinmasi gerekir (logout/login veya token refresh).

## Role Atama

Bir kullaniciya rol eklemek:

```sql
insert into public.user_roles (user_id, role)
values ('USER_UUID_HERE'::uuid, 'editor')
on conflict (user_id, role) do nothing;
```

## Profiles Akisi

- `auth.users` insert oldugunda trigger `public.profiles` satiri olusturur.
- `profiles.is_anonymous`, `auth.users.is_anonymous` ile doldurulur/guncellenir.
- `display_name`, `avatar_url` gibi app alanlari uygulama tarafindan update edilir.

## PostgreSQL (Supabase) Veri Modeli

- `auth.users`: Supabase Auth'in sistem tablosu (kayit/login kimlik kaynagi)
- `public.profiles`: uygulama profil verisi (`auth.users.id` ile birebir)
- `public.user_roles`: kullaniciya birden fazla rol atamasi icin baglanti tablosu
- RLS/policy kurallari Postgres seviyesinde uygulanir, API tarafi JWT claim ile yetkiyi dogrular

## Supabase Kurulum Komutlari

Projeyi linkle:

```bash
npx supabase link --project-ref <PROJECT_REF>
```

Migrationlari uygula:

```bash
npx supabase db push --linked
```

Edge function deploy:

```bash
npx supabase functions deploy register-push-token --project-ref <PROJECT_REF>
```

## Local Calistirma

```bash
gosupabase dev
```

veya:

```bash
go run ./cmd/api
```

## Notlar

- `supabase/migrations` altinda bazi migrationlar gecis/legacy amacli tutuluyor.
- Kaynak role modeli olarak `public.user_roles` kabul edilmelidir.
- `public.users` yeni gelistirmede kullanilmaz; profile icin `public.profiles` kullanilir.
