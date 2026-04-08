begin;

alter table if exists public.profiles
  add column if not exists is_anonymous boolean not null default false;

create or replace function public.on_auth_user_created()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into public.profiles (id, is_anonymous)
  values (new.id, coalesce(new.is_anonymous, false))
  on conflict (id) do update
    set is_anonymous = coalesce(excluded.is_anonymous, public.profiles.is_anonymous);
  return new;
end;
$$;

-- Backfill: mevcut profiller için auth.users.is_anonymous değerini eşitle
update public.profiles p
set is_anonymous = coalesce(au.is_anonymous, false)
from auth.users au
where au.id = p.id;

commit;
