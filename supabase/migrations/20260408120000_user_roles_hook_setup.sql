begin;

do $$
begin
  create type public.app_role as enum ('editor', 'admin', 'moderator');
exception
  when duplicate_object then null;
end $$;

create table if not exists public.user_roles (
  id bigint generated always as identity primary key,
  user_id uuid not null references auth.users (id) on delete cascade,
  role public.app_role not null,
  unique (user_id, role)
);

grant usage on schema public to supabase_auth_admin;
grant select on table public.user_roles to supabase_auth_admin;
revoke all on table public.user_roles from authenticated, anon, public;

alter table public.user_roles enable row level security;

drop policy if exists "auth admin read user_roles" on public.user_roles;
create policy "auth admin read user_roles"
  on public.user_roles
  as permissive
  for select
  to supabase_auth_admin
  using (true);

create or replace function public.custom_access_token_hook(event jsonb)
returns jsonb
language plpgsql
stable
security definer
set search_path = public
as $function$
declare
  claims     jsonb;
  roles_json jsonb;
  uid        uuid;
begin
  uid := (event->>'user_id')::uuid;

  select coalesce(
    (select jsonb_agg(ur.role::text order by ur.role)
     from public.user_roles ur
     where ur.user_id = uid),
    '[]'::jsonb
  ) into roles_json;

  claims := coalesce(event->'claims', '{}'::jsonb);
  claims := jsonb_set(claims, '{roles}', roles_json, true);

  event := jsonb_set(event, '{claims}', claims, true);
  return event;
end;
$function$;

revoke all on function public.custom_access_token_hook(jsonb) from public;
grant execute on function public.custom_access_token_hook(jsonb) to supabase_auth_admin;

commit;
