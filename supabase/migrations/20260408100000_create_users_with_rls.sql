create table if not exists public.users (
  id uuid primary key references auth.users (id) on delete cascade,
  push_token text,
  language text,
  push_token_updated_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint users_language_check check (
    language is null or language in ('tr', 'en', 'fr', 'de', 'es', 'ar')
  )
);

alter table public.users enable row level security;

drop policy if exists "users_select_own" on public.users;
create policy "users_select_own"
  on public.users
  for select
  to authenticated
  using (auth.uid() = id);

drop policy if exists "users_update_own" on public.users;
create policy "users_update_own"
  on public.users
  for update
  to authenticated
  using (auth.uid() = id)
  with check (auth.uid() = id);
