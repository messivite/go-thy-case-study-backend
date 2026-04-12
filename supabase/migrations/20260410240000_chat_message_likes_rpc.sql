-- Kişisel mesaj beğenileri: oturum sahibi yalnızca kendi user mesajları ve asistan yanıtlarını işaretler.
-- Tek RPC: doğrulama + insert/delete + güncel state (1=liked, 2=unliked) — tek round-trip.

create table if not exists public.chat_message_likes (
    user_id    uuid not null references auth.users (id) on delete cascade,
    message_id uuid not null references public.chat_messages (id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (user_id, message_id)
);

create index if not exists idx_chat_message_likes_message_id
    on public.chat_message_likes (message_id);

comment on table public.chat_message_likes is
    'Kullanıcı başına mesaj başına en fazla bir beğeni; sadece oturum sahibi, kendi user mesajı veya asistan yanıtı.';

alter table public.chat_message_likes enable row level security;

-- Doğrudan istemci erişimi yok; yalnızca service_role (API) ve security definer RPC yazar.
revoke all on public.chat_message_likes from public;
grant select, insert, delete on public.chat_message_likes to service_role;

create or replace function public.set_chat_message_like(
    p_user_id uuid,
    p_session_id uuid,
    p_message_id uuid,
    p_action smallint
)
returns jsonb
language plpgsql
volatile
security definer
set search_path = public
as $$
declare
    v_owner       uuid;
    v_sess_del    timestamptz;
    v_role        text;
    v_msg_user    uuid;
    v_msg_deleted timestamptz;
    v_sess_of_msg uuid;
    v_liked       boolean;
begin
    if p_action is null or p_action not in (1, 2) then
        raise exception 'invalid_action';
    end if;

    select s.user_id, s.deleted_at
    into v_owner, v_sess_del
    from public.chat_sessions s
    where s.id = p_session_id;

    if v_owner is null or v_sess_del is not null then
        raise exception 'session_not_found';
    end if;

    if v_owner <> p_user_id then
        raise exception 'unauthorized';
    end if;

    select m.session_id, m.role, m.user_id, m.deleted_at
    into v_sess_of_msg, v_role, v_msg_user, v_msg_deleted
    from public.chat_messages m
    where m.id = p_message_id;

    if v_sess_of_msg is null or v_sess_of_msg <> p_session_id then
        raise exception 'message_not_found';
    end if;

    if v_msg_deleted is not null then
        raise exception 'message_not_found';
    end if;

    if v_role = 'system' then
        raise exception 'message_not_likeable';
    end if;

    if v_role = 'user' then
        if v_msg_user is null or v_msg_user <> p_user_id then
            raise exception 'message_not_likeable';
        end if;
    end if;

    if p_action = 1 then
        insert into public.chat_message_likes (user_id, message_id)
        values (p_user_id, p_message_id)
        on conflict (user_id, message_id) do nothing;
    else
        delete from public.chat_message_likes
        where user_id = p_user_id
          and message_id = p_message_id;
    end if;

    select exists(
        select 1
        from public.chat_message_likes l
        where l.user_id = p_user_id
          and l.message_id = p_message_id
    )
    into v_liked;

    if v_liked then
        return jsonb_build_object('state', 1);
    end if;

    return jsonb_build_object('state', 2);
end;
$$;

comment on function public.set_chat_message_like(uuid, uuid, uuid, smallint) is
    'p_action: 1=like (insert), 2=unlike (delete). Dönüş: {"state":1} beğenildi, {"state":2} beğenilmedi.';

revoke all on function public.set_chat_message_like(uuid, uuid, uuid, smallint) from public;
grant execute on function public.set_chat_message_like(uuid, uuid, uuid, smallint) to service_role;
