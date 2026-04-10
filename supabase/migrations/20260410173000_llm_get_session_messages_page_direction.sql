-- llm_get_session_messages_page fonksiyonunu direction (older/newer) desteğiyle günceller.
create or replace function public.llm_get_session_messages_page(
    p_session_id uuid,
    p_limit integer default 50,
    p_direction text default 'older',
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
ordered as (
    select *
    from windowed w
    where
        (p_cursor_created_at is null and p_cursor_message_id is null)
        or (
            p_direction = 'older'
            and (w.created_at, w.id) < (p_cursor_created_at, p_cursor_message_id)
        )
        or (
            p_direction = 'newer'
            and (w.created_at, w.id) > (p_cursor_created_at, p_cursor_message_id)
        )
    order by
        case when p_direction = 'newer' then w.created_at end asc,
        case when p_direction = 'newer' then w.id end asc,
        case when p_direction <> 'newer' then w.created_at end desc,
        case when p_direction <> 'newer' then w.id end desc
    limit greatest(1, least(coalesce(p_limit, 50), 100))
)
select
    o.total_count_all as total_count,
    o.id,
    o.session_id,
    o.user_id,
    o.role,
    o.content,
    o.created_at,
    o.provider,
    o.model
from ordered o;
$$;

revoke all on function public.llm_get_session_messages_page(uuid, integer, text, timestamptz, uuid) from public;
grant execute on function public.llm_get_session_messages_page(uuid, integer, text, timestamptz, uuid) to service_role;
