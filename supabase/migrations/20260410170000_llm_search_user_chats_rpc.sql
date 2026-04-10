-- Kullanıcının chat başlıkları ve user/assistant mesaj içeriklerinde arama.
-- Cursor yapısı: sort_at DESC, session_id DESC (keyset pagination).
create or replace function public.llm_search_user_chats(
    p_user_id uuid,
    p_query text,
    p_limit integer default 20,
    p_cursor_sort_at timestamptz default null,
    p_cursor_session_id uuid default null
)
returns table (
    total_count integer,
    session_id uuid,
    title text,
    session_created_at timestamptz,
    session_updated_at timestamptz,
    last_message_at timestamptz,
    title_matched boolean,
    matched_message_id uuid,
    matched_role text,
    matched_content text,
    matched_at timestamptz,
    sort_at timestamptz
)
language sql
stable
security definer
set search_path = public
as $$
with matched as (
    select
        s.id as session_id,
        s.title,
        s.created_at as session_created_at,
        s.updated_at as session_updated_at,
        lm.last_message_at,
        (position(lower(trim(p_query)) in lower(s.title)) > 0) as title_matched,
        mm.id as matched_message_id,
        mm.role as matched_role,
        mm.content as matched_content,
        mm.created_at as matched_at,
        coalesce(mm.created_at, lm.last_message_at, s.updated_at, s.created_at) as sort_at
    from public.chat_sessions s
    left join lateral (
        select max(m.created_at) as last_message_at
        from public.chat_messages m
        where m.session_id = s.id
    ) lm on true
    left join lateral (
        select m.id, m.role, m.content, m.created_at
        from public.chat_messages m
        where m.session_id = s.id
          and m.role in ('user', 'assistant')
          and position(lower(trim(p_query)) in lower(m.content)) > 0
        order by m.created_at desc, m.id desc
        limit 1
    ) mm on true
    where s.user_id = p_user_id
      and (
        position(lower(trim(p_query)) in lower(s.title)) > 0
        or mm.id is not null
      )
),
windowed as (
    select m.*, count(*) over()::integer as total_count_all
    from matched m
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
    p.session_created_at,
    p.session_updated_at,
    p.last_message_at,
    p.title_matched,
    p.matched_message_id,
    p.matched_role,
    p.matched_content,
    p.matched_at,
    p.sort_at
from paged p;
$$;

revoke all on function public.llm_search_user_chats(uuid, text, integer, timestamptz, uuid) from public;
grant execute on function public.llm_search_user_chats(uuid, text, integer, timestamptz, uuid) to service_role;
