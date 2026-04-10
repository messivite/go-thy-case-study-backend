-- Uygulama açılışında API tarafından doldurulan desteklenen LLM model kataloğu.
-- İstemci listesi: GET /api/models (is_active = true).
-- Koddan kalkan modeller sync ile is_active = false olur; operatör isterse SQL ile is_active kapatabilir
-- (bir sonraki sync, katalogda hâlâ varsa tekrar true yapar — tam kapatmak için koddan kaldırın).

begin;

create table if not exists public.llm_supported_models (
    provider text not null,
    model_id text not null,
    display_name text not null default '',
    supports_stream boolean not null default true,
    is_active boolean not null default true,
    updated_at timestamptz not null default now(),
    primary key (provider, model_id)
);

create index if not exists llm_supported_models_active_list_idx
    on public.llm_supported_models (provider, model_id)
    where is_active = true;

alter table public.llm_supported_models enable row level security;

drop policy if exists "llm_supported_models_select_active" on public.llm_supported_models;
create policy "llm_supported_models_select_active"
    on public.llm_supported_models
    for select
    to authenticated
    using (is_active = true);

-- service_role PostgREST ile RLS bypass; doğrudan Supabase client kullanan uygulamalar için yukarıdaki politika.

create or replace function public.llm_sync_supported_models(p_payload jsonb)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
    if p_payload is null or jsonb_typeof(p_payload) != 'array' then
        raise exception 'p_payload must be a json array';
    end if;

    insert into public.llm_supported_models (provider, model_id, display_name, supports_stream, is_active, updated_at)
    select
        (e->>'provider')::text,
        (e->>'model_id')::text,
        coalesce(nullif(trim(e->>'display_name'), ''), (e->>'model_id')::text),
        coalesce((e->>'supports_stream')::boolean, true),
        true,
        now()
    from jsonb_array_elements(p_payload) as e
    on conflict (provider, model_id) do update set
        display_name = excluded.display_name,
        supports_stream = excluded.supports_stream,
        is_active = true,
        updated_at = now();

    update public.llm_supported_models o
    set is_active = false,
        updated_at = now()
    where not exists (
        select 1
        from jsonb_array_elements(p_payload) e
        where (e->>'provider')::text = o.provider
          and (e->>'model_id')::text = o.model_id
    );
end;
$$;

revoke all on function public.llm_sync_supported_models(jsonb) from public;
grant execute on function public.llm_sync_supported_models(jsonb) to service_role;

commit;
