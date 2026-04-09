-- Varsayılan kota ayarları (singleton satır).
create table if not exists public.llm_quota_defaults (
    id                     text primary key default 'default'
        check (id = 'default'),
    default_daily_tokens   integer not null default 100000,
    default_weekly_tokens  integer not null default 500000,
    period_timezone        text    not null default 'UTC',
    created_at             timestamptz not null default now(),
    updated_at             timestamptz not null default now()
);

comment on table public.llm_quota_defaults is
    'Tekil satır: yeni kullanıcılara atanacak varsayılan günlük/haftalık token limiti.';

alter table public.llm_quota_defaults enable row level security;

insert into public.llm_quota_defaults (id) values ('default')
on conflict (id) do nothing;

-- Kullanıcıya özel kota (PK = user_id; her profil için trigger ile üretilir).
create table if not exists public.user_llm_usage_quota (
    user_id               uuid primary key references auth.users (id) on delete cascade,
    daily_token_limit     integer not null,
    weekly_token_limit    integer not null,
    quota_bypass          boolean not null default false,
    created_at            timestamptz not null default now(),
    updated_at            timestamptz not null default now()
);

comment on table public.user_llm_usage_quota is
    'Kullanıcı başına 1 satır kota: günlük/haftalık limit ve admin bypass.';
comment on column public.user_llm_usage_quota.quota_bypass is
    'true iken Go API limit kontrolünü atlar (admin / destek override).';

alter table public.user_llm_usage_quota enable row level security;

create index if not exists idx_user_llm_usage_quota_bypass
    on public.user_llm_usage_quota (quota_bypass)
    where quota_bypass = true;

-- updated_at trigger (update_updated_at fonksiyonu zaten mevcut).
create trigger llm_quota_defaults_updated_at
    before update on public.llm_quota_defaults
    for each row
    execute function public.update_updated_at();

create trigger user_llm_usage_quota_updated_at
    before update on public.user_llm_usage_quota
    for each row
    execute function public.update_updated_at();

-- Profil oluşunca default limitlerle kota satırı üret.
create or replace function public.trg_create_user_quota()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
    insert into public.user_llm_usage_quota (
        user_id, daily_token_limit, weekly_token_limit, quota_bypass
    )
    select
        NEW.id,
        coalesce(d.default_daily_tokens, 100000),
        coalesce(d.default_weekly_tokens, 500000),
        false
    from public.llm_quota_defaults d
    where d.id = 'default'
    on conflict (user_id) do nothing;

    return NEW;
end;
$$;

drop trigger if exists trg_profiles_create_quota on public.profiles;
create trigger trg_profiles_create_quota
    after insert on public.profiles
    for each row
    execute function public.trg_create_user_quota();

-- Mevcut kullanıcılar için backfill.
insert into public.user_llm_usage_quota (user_id, daily_token_limit, weekly_token_limit, quota_bypass)
select
    p.id,
    coalesce(d.default_daily_tokens, 100000),
    coalesce(d.default_weekly_tokens, 500000),
    false
from public.profiles p
cross join public.llm_quota_defaults d
where d.id = 'default'
  and not exists (
      select 1 from public.user_llm_usage_quota q where q.user_id = p.id
  );
