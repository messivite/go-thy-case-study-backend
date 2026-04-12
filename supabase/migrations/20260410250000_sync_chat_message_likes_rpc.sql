-- Offline like/unlike senkronu: tek RPC ile 1..N öğe; oturum doğrulaması bir kez, satır bazlı hata kodları.

create or replace function public.sync_chat_message_likes(
    p_user_id uuid,
    p_session_id uuid,
    p_items jsonb
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
    v_len         int;
    i             int;
    elem          jsonb;
    v_mid_text    text;
    v_mid         uuid;
    v_action_text text;
    v_action      int;
    v_results     jsonb := '[]'::jsonb;
    v_row         jsonb;
    v_sess_of_msg uuid;
    v_role        text;
    v_msg_user    uuid;
    v_msg_deleted timestamptz;
    v_liked       boolean;
begin
    if p_items is null or jsonb_typeof(p_items) <> 'array' then
        return jsonb_build_object('error', 'invalid_items');
    end if;

    v_len := jsonb_array_length(p_items);
    if v_len < 1 then
        return jsonb_build_object('error', 'empty_items');
    end if;
    if v_len > 100 then
        return jsonb_build_object('error', 'too_many_items');
    end if;

    select s.user_id, s.deleted_at
    into v_owner, v_sess_del
    from public.chat_sessions s
    where s.id = p_session_id;

    if v_owner is null or v_sess_del is not null then
        return jsonb_build_object('error', 'session_not_found');
    end if;

    if v_owner <> p_user_id then
        return jsonb_build_object('error', 'unauthorized');
    end if;

    for i in 0 .. v_len - 1 loop
        elem := p_items -> i;
        v_mid_text := coalesce(nullif(trim(elem ->> 'messageId'), ''), nullif(trim(elem ->> 'message_id'), ''));
        v_action_text := nullif(trim(elem ->> 'action'), '');

        if v_mid_text is null or v_action_text is null then
            v_row := jsonb_build_object(
                'messageId', coalesce(v_mid_text, ''),
                'ok', false,
                'code', 'invalid_item'
            );
            v_results := v_results || jsonb_build_array(v_row);
            continue;
        end if;

        begin
            v_mid := v_mid_text::uuid;
        exception
            when invalid_text_representation then
                v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'invalid_message_id');
                v_results := v_results || jsonb_build_array(v_row);
                continue;
        end;

        begin
            v_action := v_action_text::int;
        exception
            when others then
                v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'invalid_action');
                v_results := v_results || jsonb_build_array(v_row);
                continue;
        end;

        if v_action not in (1, 2) then
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'invalid_action');
            v_results := v_results || jsonb_build_array(v_row);
            continue;
        end if;

        select m.session_id, m.role, m.user_id, m.deleted_at
        into v_sess_of_msg, v_role, v_msg_user, v_msg_deleted
        from public.chat_messages m
        where m.id = v_mid;

        if v_sess_of_msg is null or v_sess_of_msg <> p_session_id then
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'message_not_found');
            v_results := v_results || jsonb_build_array(v_row);
            continue;
        end if;

        if v_msg_deleted is not null then
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'message_not_found');
            v_results := v_results || jsonb_build_array(v_row);
            continue;
        end if;

        if v_role = 'system' then
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'message_not_likeable');
            v_results := v_results || jsonb_build_array(v_row);
            continue;
        end if;

        if v_role = 'user' then
            if v_msg_user is null or v_msg_user <> p_user_id then
                v_row := jsonb_build_object('messageId', v_mid_text, 'ok', false, 'code', 'message_not_likeable');
                v_results := v_results || jsonb_build_array(v_row);
                continue;
            end if;
        end if;

        if v_action = 1 then
            insert into public.chat_message_likes (user_id, message_id)
            values (p_user_id, v_mid)
            on conflict (user_id, message_id) do nothing;
        else
            delete from public.chat_message_likes
            where user_id = p_user_id
              and message_id = v_mid;
        end if;

        select exists(
            select 1
            from public.chat_message_likes l
            where l.user_id = p_user_id
              and l.message_id = v_mid
        )
        into v_liked;

        if v_liked then
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', true, 'state', 1);
        else
            v_row := jsonb_build_object('messageId', v_mid_text, 'ok', true, 'state', 2);
        end if;
        v_results := v_results || jsonb_build_array(v_row);
    end loop;

    return jsonb_build_object('results', v_results);
end;
$$;

comment on function public.sync_chat_message_likes(uuid, uuid, jsonb) is
    'p_items: JSON dizi [{ "messageId": uuid, "action": 1|2 }, ...]; en fazla 100. Dönüş: { "results": [...] } veya { "error": "..." }.';

revoke all on function public.sync_chat_message_likes(uuid, uuid, jsonb) from public;
grant execute on function public.sync_chat_message_likes(uuid, uuid, jsonb) to service_role;
