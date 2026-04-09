alter table public.llm_interaction_log
    drop constraint if exists llm_interaction_log_outcome_check;

alter table public.llm_interaction_log
    add constraint llm_interaction_log_outcome_check
    check (outcome in ('pending', 'ok', 'error', 'cancelled'));

create or replace function public.llm_cancel_pending_for_user_message(
    p_user_message_id uuid
)
returns void
language plpgsql
security definer
set search_path = public
as $$
begin
    update public.llm_interaction_log
    set
        outcome = 'cancelled',
        error_summary = 'generation cancelled by user',
        error_code = 'user_cancelled',
        updated_at = now()
    where user_message_id = p_user_message_id
      and outcome = 'pending';
end;
$$;

comment on function public.llm_cancel_pending_for_user_message(uuid) is
    'Service role / API: marks pending llm interaction as cancelled by user.';

revoke all on function public.llm_cancel_pending_for_user_message(uuid) from public;
grant execute on function public.llm_cancel_pending_for_user_message(uuid) to service_role;
