begin;

create or replace function public.sync_auth_user_roles(target_user_id uuid)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
  update auth.users au
  set raw_app_meta_data = jsonb_set(
    coalesce(au.raw_app_meta_data, '{}'::jsonb),
    '{roles}',
    (
      select coalesce(jsonb_agg(ur.role::text order by ur.role), '[]'::jsonb)
      from public.user_roles ur
      where ur.user_id = target_user_id
    ),
    true
  )
  where au.id = target_user_id;
end;
$$;

create or replace function public.trg_sync_auth_user_roles()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  if tg_op = 'INSERT' then
    perform public.sync_auth_user_roles(new.user_id);
    return new;
  elsif tg_op = 'UPDATE' then
    if old.user_id is distinct from new.user_id then
      perform public.sync_auth_user_roles(old.user_id);
    end if;
    perform public.sync_auth_user_roles(new.user_id);
    return new;
  elsif tg_op = 'DELETE' then
    perform public.sync_auth_user_roles(old.user_id);
    return old;
  end if;
  return null;
end;
$$;

drop trigger if exists trg_user_roles_sync_auth_metadata on public.user_roles;
create trigger trg_user_roles_sync_auth_metadata
after insert or update or delete on public.user_roles
for each row
execute function public.trg_sync_auth_user_roles();

-- Backfill: mevcut role verilerini auth.users.raw_app_meta_data.roles ile hizala
update auth.users au
set raw_app_meta_data = jsonb_set(
  coalesce(au.raw_app_meta_data, '{}'::jsonb),
  '{roles}',
  (
    select coalesce(jsonb_agg(ur.role::text order by ur.role), '[]'::jsonb)
    from public.user_roles ur
    where ur.user_id = au.id
  ),
  true
);

commit;
