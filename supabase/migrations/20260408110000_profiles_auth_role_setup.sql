begin;

-- Legacy users modelini kapat (varsa)
drop trigger if exists on_auth_user_created on auth.users;
drop function if exists public.on_auth_user_created();
drop table if exists public.users cascade;

-- profiles tablosu (auth.users ile 1:1)
create table if not exists public.profiles (
  id uuid primary key references auth.users (id) on delete cascade,
  display_name text null,
  avatar_url text null,
  role text not null default 'user',
  is_active boolean not null default true,
  preferred_provider text null,
  preferred_model text null,
  locale text not null default 'tr',
  timezone text null,
  metadata jsonb not null default '{}'::jsonb,
  last_seen_at timestamptz null,
  onboarding_completed boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint profiles_role_check check (role in ('user', 'admin', 'moderator'))
);

create index if not exists profiles_role_idx on public.profiles (role);

-- updated_at otomatik güncelleme
create or replace function public.set_updated_at()
returns trigger
language plpgsql
security invoker
set search_path = public
as $$
begin
  new.updated_at := now();
  return new;
end;
$$;

drop trigger if exists trg_profiles_set_updated_at on public.profiles;
create trigger trg_profiles_set_updated_at
before update on public.profiles
for each row
execute function public.set_updated_at();

-- Yeni auth user -> profiles satırı
create or replace function public.on_auth_user_created()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into public.profiles (id)
  values (new.id)
  on conflict (id) do nothing;
  return new;
end;
$$;

drop trigger if exists on_auth_user_created on auth.users;
create trigger on_auth_user_created
after insert on auth.users
for each row
execute function public.on_auth_user_created();

-- RLS + policy
alter table public.profiles enable row level security;

drop policy if exists "profiles_select_own" on public.profiles;
create policy "profiles_select_own"
on public.profiles
for select
to authenticated
using (auth.uid() = id);

drop policy if exists "profiles_update_own" on public.profiles;
create policy "profiles_update_own"
on public.profiles
for update
to authenticated
using (auth.uid() = id)
with check (auth.uid() = id);

drop policy if exists "profiles_insert_own" on public.profiles;
create policy "profiles_insert_own"
on public.profiles
for insert
to authenticated
with check (auth.uid() = id);

-- Access token hook (profiles.role -> claims.app_role)
create or replace function public.custom_access_token_hook(event jsonb)
returns jsonb
language plpgsql
stable
security definer
set search_path = public
as $$
declare
  claims jsonb;
  uid uuid;
  user_role text;
begin
  uid := nullif(event->>'user_id', '')::uuid;

  select p.role
    into user_role
  from public.profiles p
  where p.id = uid;

  if user_role is null then
    user_role := 'user';
  end if;

  claims := coalesce(event->'claims', '{}'::jsonb);
  claims := jsonb_set(claims, '{app_role}', to_jsonb(user_role), true);
  event := jsonb_set(event, '{claims}', claims, true);

  return event;
end;
$$;

revoke all on function public.custom_access_token_hook(jsonb) from public;
grant execute on function public.custom_access_token_hook(jsonb) to supabase_auth_admin;

commit;
