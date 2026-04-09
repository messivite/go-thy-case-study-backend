-- LLM etkileşim günlüğü: chat_messages üzerinde AFTER INSERT trigger ile otomatik satır.
-- Sınırlar: LLM hatasında asistan satırı hiç INSERT edilmezse trigger tetiklenmez; bu durumda
-- API service role ile public.llm_fail_pending_for_user_message veya usage için
-- public.llm_set_usage_for_user_message çağırmalıdır.

create table if not exists public.llm_interaction_log (
    id                      uuid primary key default gen_random_uuid(),
    session_id              uuid not null references public.chat_sessions (id) on delete cascade,
    user_id                 uuid not null references auth.users (id) on delete cascade,
    user_message_id         uuid not null references public.chat_messages (id) on delete cascade,
    user_sent_at            timestamptz not null,
    user_prompt_excerpt     text        not null default '',
    assistant_message_id    uuid references public.chat_messages (id) on delete set null,
    assistant_received_at   timestamptz,
    assistant_excerpt       text,
    provider                text,
    model                   text,
    outcome                 text        not null default 'pending'
        check (outcome in ('pending', 'ok', 'error')),
    error_summary           text,
    prompt_tokens           integer,
    completion_tokens       integer,
    total_tokens            integer,
    created_at              timestamptz not null default now(),
    updated_at              timestamptz not null default now(),
    constraint llm_interaction_log_user_message_unique unique (user_message_id)
);

comment on table public.llm_interaction_log is
    'Kullanıcı mesajı ve (varsa) asistan cevabı çifti; raporlama için. user satırı INSERT ile trigger yazar; assistant INSERT ile tamamlanır; hata ve token alanları API/RPC ile doldurulur.';

comment on column public.llm_interaction_log.outcome is
    'pending: asistan bekleniyor; ok: trigger ile assistant satırı eşlendi; error: API llm_fail_pending_for_user_message ile işaretledi.';
comment on column public.llm_interaction_log.user_prompt_excerpt is
    'İçeriğin ilk 500 karakteri (trigger). Tam metin chat_messages.content.';

create index if not exists idx_llm_interaction_log_user_sent
    on public.llm_interaction_log (user_id, user_sent_at desc);

create index if not exists idx_llm_interaction_log_session
    on public.llm_interaction_log (session_id, user_sent_at desc);

create index if not exists idx_llm_interaction_log_outcome
    on public.llm_interaction_log (outcome, user_sent_at desc)
    where outcome = 'pending';

-- RLS: JWT ile doğrudan erişim yok; service_role RLS bypass eder (Go API).
alter table public.llm_interaction_log enable row level security;

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
    if TG_OP <> 'INSERT' then
        return NEW;
    end if;

    if NEW.role = 'user' then
        select cs.user_id
        into v_user_id
        from public.chat_sessions cs
        where cs.id = NEW.session_id;

        if v_user_id is null then
            return NEW;
        end if;

        insert into public.llm_interaction_log (
            session_id,
            user_id,
            user_message_id,
            user_sent_at,
            user_prompt_excerpt,
            outcome
        ) values (
            NEW.session_id,
            v_user_id,
            NEW.id,
            NEW.created_at,
            left(coalesce(NEW.content, ''), 500),
            'pending'
        );

    elsif NEW.role = 'assistant' then
        update public.llm_interaction_log l
        set
            assistant_message_id  = NEW.id,
            assistant_received_at = NEW.created_at,
            assistant_excerpt     = left(coalesce(NEW.content, ''), 500),
            provider              = NEW.provider,
            model                 = NEW.model,
            outcome               = 'ok',
            updated_at            = now()
        where l.id = (
            select l2.id
            from public.llm_interaction_log l2
            where l2.session_id = NEW.session_id
              and l2.outcome = 'pending'
            order by l2.user_sent_at asc
            limit 1
        )
        returning l.id into v_updated;
    end if;

    return NEW;
end;
$$;

drop trigger if exists chat_messages_llm_audit on public.chat_messages;
create trigger chat_messages_llm_audit
    after insert on public.chat_messages
    for each row
    execute function public.trg_chat_messages_llm_audit();

create trigger llm_interaction_log_updated_at
    before update on public.llm_interaction_log
    for each row
    execute function public.update_updated_at();

-- LLM hatası (asistan satırı yok): en son pending satırı user_message_id ile kapat.
create or replace function public.llm_fail_pending_for_user_message(
    p_user_message_id uuid,
    p_error_summary text
)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
    update public.llm_interaction_log
    set
        outcome         = 'error',
        error_summary   = left(coalesce(p_error_summary, ''), 2000),
        updated_at      = now()
    where user_message_id = p_user_message_id
      and outcome = 'pending';
end;
$$;

comment on function public.llm_fail_pending_for_user_message(uuid, text) is
    'Service role / API: LLM veya upstream hata; ilgili kullanıcı mesajına ait pending log satırını error yapar.';

-- Opsiyonel: completion sonrası token sayıları (API OpenAI usage ile doldurur).
create or replace function public.llm_set_usage_for_user_message(
    p_user_message_id uuid,
    p_prompt_tokens integer,
    p_completion_tokens integer,
    p_total_tokens integer
)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
    update public.llm_interaction_log
    set
        prompt_tokens     = p_prompt_tokens,
        completion_tokens = p_completion_tokens,
        total_tokens      = coalesce(p_total_tokens, p_prompt_tokens + coalesce(p_completion_tokens, 0)),
        updated_at        = now()
    where user_message_id = p_user_message_id;
end;
$$;

comment on function public.llm_set_usage_for_user_message(uuid, integer, integer, integer) is
    'Service role / API: ilgili etkileşim satırına usage yazar (outcome pending veya ok iken).';

revoke all on function public.llm_fail_pending_for_user_message(uuid, text) from public;
revoke all on function public.llm_set_usage_for_user_message(uuid, integer, integer, integer) from public;

grant execute on function public.llm_fail_pending_for_user_message(uuid, text) to service_role;
grant execute on function public.llm_set_usage_for_user_message(uuid, integer, integer, integer) to service_role;
