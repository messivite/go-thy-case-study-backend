-- Chat mesajları için cursor-based sayfalama (newest -> oldest), total_count dahil.
create or replace function public.llm_get_session_messages_page(
    p_session_id uuid,
    p_limit integer default 50,
    p_cursor_created_at timestamptz default null,
    p_cursor_message_id uuid default null
)
returns table (
    total_count integer,
    id uuid,
    session_id uuid,
    user_id uuid,
    role text,
    content text,
    created_at timestamptz,
    provider text,
    model text
)
language sql
stable
security definer
set search_path = public
as $$
with base as (
    select m.*
    from public.chat_messages m
    where m.session_id = p_session_id
),
windowed as (
    select b.*, count(*) over()::integer as total_count_all
    from base b
),
paged as (
    select *
    from windowed w
    where
        (p_cursor_created_at is null and p_cursor_message_id is null)
        or (w.created_at, w.id) < (p_cursor_created_at, p_cursor_message_id)
    order by w.created_at desc, w.id desc
    limit greatest(1, least(coalesce(p_limit, 50), 100))
)
select
    p.total_count_all as total_count,
    p.id,
    p.session_id,
    p.user_id,
    p.role,
    p.content,
    p.created_at,
    p.provider,
    p.model
from paged p;
$$;

revoke all on function public.llm_get_session_messages_page(uuid, integer, timestamptz, uuid) from public;
grant execute on function public.llm_get_session_messages_page(uuid, integer, timestamptz, uuid) to service_role;
