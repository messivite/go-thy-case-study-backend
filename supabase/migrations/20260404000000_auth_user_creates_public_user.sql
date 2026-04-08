-- auth.users'a her yeni kullanıcı eklendiğinde public.users'ta aynı id ile satır oluşturur.
-- public.users'ta id dışındaki kolonlar NULL / default kabul edilebilir olmalı (veya aşağıdaki INSERT'i genişlet).

create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into public.users (id)
  values (new.id)
  on conflict (id) do nothing;
  return new;
end;
$$;

drop trigger if exists on_auth_user_created on auth.users;

create trigger on_auth_user_created
  after insert on auth.users
  for each row
  execute function public.handle_new_user();

comment on function public.handle_new_user() is 'Mirrors new auth.users row into public.users (id).';
