-- Chat listesi için cursor-based sayfalama (sort_at DESC, session_id DESC).
create or replace function public.llm_get_user_chat_sessions_page(
    p_user_id uuid,
    p_limit integer default 20,
    p_cursor_sort_at timestamptz default null,
    p_cursor_session_id uuid default null
)
returns table (
    total_count integer,
    session_id uuid,
    title text,
    created_at timestamptz,
    updated_at timestamptz,
    default_provider text,
    default_model text,
    last_provider text,
    last_model text,
    last_message_preview text,
    sort_at timestamptz
)
language sql
stable
security definer
set search_path = public
as $$
with base as (
    select
        s.id as session_id,
        s.title,
        s.created_at,
        s.updated_at,
        s.default_provider,
        s.default_model,
        s.last_provider,
        s.last_model,
        lm.last_content as last_message_preview,
        coalesce(lm.last_created_at, s.updated_at, s.created_at) as sort_at
    from public.chat_sessions s
    left join lateral (
        select
            m.content as last_content,
            m.created_at as last_created_at
        from public.chat_messages m
        where m.session_id = s.id
        order by m.created_at desc, m.id desc
        limit 1
    ) lm on true
    where s.user_id = p_user_id
),
windowed as (
    select b.*, count(*) over()::integer as total_count_all
    from base b
),
paged as (
    select *
    from windowed w
    where
        (p_cursor_sort_at is null and p_cursor_session_id is null)
        or (w.sort_at, w.session_id) < (p_cursor_sort_at, p_cursor_session_id)
    order by w.sort_at desc, w.session_id desc
    limit greatest(1, least(coalesce(p_limit, 20), 100))
)
select
    p.total_count_all as total_count,
    p.session_id,
    p.title,
    p.created_at,
    p.updated_at,
    p.default_provider,
    p.default_model,
    p.last_provider,
    p.last_model,
    p.last_message_preview,
    p.sort_at
from paged p;
$$;

revoke all on function public.llm_get_user_chat_sessions_page(uuid, integer, timestamptz, uuid) from public;
grant execute on function public.llm_get_user_chat_sessions_page(uuid, integer, timestamptz, uuid) to service_role;
