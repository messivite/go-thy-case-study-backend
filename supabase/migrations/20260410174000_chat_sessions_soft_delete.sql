-- chat_sessions soft-delete alanı
alter table public.chat_sessions
    add column if not exists deleted_at timestamptz null;

create index if not exists idx_chat_sessions_user_not_deleted
    on public.chat_sessions (user_id, created_at desc)
    where deleted_at is null;

comment on column public.chat_sessions.deleted_at is
    'Soft delete zamanı; null ise aktif.';

-- Arama RPC: silinmiş oturumları dışla
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
      and s.deleted_at is null
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

-- Chat list RPC: silinmiş oturumları dışla
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
      and s.deleted_at is null
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

-- Message page RPC: silinmiş oturumlar için sonuç döndürme
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
      and exists (
        select 1
        from public.chat_sessions s
        where s.id = p_session_id
          and s.deleted_at is null
      )
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
