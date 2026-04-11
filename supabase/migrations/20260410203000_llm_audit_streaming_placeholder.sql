-- Streaming: assistant row may be INSERTed with empty content first; complete llm_interaction_log
-- when content becomes non-empty (UPDATE), or on INSERT when content is already non-empty.

create or replace function public.trg_chat_messages_llm_audit()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
declare
    v_user_id uuid;
    v_updated uuid;
begin
    if tg_op = 'INSERT' and new.role = 'user' then
        select cs.user_id
        into v_user_id
        from public.chat_sessions cs
        where cs.id = new.session_id;

        if v_user_id is null then
            return new;
        end if;

        insert into public.llm_interaction_log (
            session_id,
            user_id,
            user_message_id,
            user_sent_at,
            user_prompt_excerpt,
            outcome
        ) values (
            new.session_id,
            v_user_id,
            new.id,
            new.created_at,
            left(coalesce(new.content, ''), 500),
            'pending'
        );

        return new;
    end if;

    if tg_op = 'INSERT' and new.role = 'assistant' then
        if trim(coalesce(new.content, '')) = '' then
            return new;
        end if;

        update public.llm_interaction_log l
        set
            assistant_message_id  = new.id,
            assistant_received_at = new.created_at,
            assistant_excerpt     = left(coalesce(new.content, ''), 500),
            provider              = new.provider,
            model                 = new.model,
            outcome               = 'ok',
            updated_at            = now()
        where l.id = (
            select l2.id
            from public.llm_interaction_log l2
            where l2.session_id = new.session_id
              and l2.outcome = 'pending'
            order by l2.user_sent_at asc
            limit 1
        )
        returning l.id into v_updated;

        return new;
    end if;

    if tg_op = 'UPDATE' and new.role = 'assistant' and old.role = 'assistant' then
        if trim(coalesce(old.content, '')) <> '' or trim(coalesce(new.content, '')) = '' then
            return new;
        end if;

        update public.llm_interaction_log l
        set
            assistant_message_id  = new.id,
            assistant_received_at = now(),
            assistant_excerpt     = left(coalesce(new.content, ''), 500),
            provider              = new.provider,
            model                 = new.model,
            outcome               = 'ok',
            updated_at            = now()
        where l.id = (
            select l2.id
            from public.llm_interaction_log l2
            where l2.session_id = new.session_id
              and l2.outcome = 'pending'
            order by l2.user_sent_at asc
            limit 1
        )
        returning l.id into v_updated;

        return new;
    end if;

    return new;
end;
$$;

drop trigger if exists chat_messages_llm_audit on public.chat_messages;

create trigger chat_messages_llm_audit
    after insert or update on public.chat_messages
    for each row
    execute function public.trg_chat_messages_llm_audit();

comment on function public.trg_chat_messages_llm_audit() is
    'User INSERT creates pending llm_interaction_log; assistant with content on INSERT or first non-empty UPDATE completes it. Empty assistant INSERT is ignored.';
