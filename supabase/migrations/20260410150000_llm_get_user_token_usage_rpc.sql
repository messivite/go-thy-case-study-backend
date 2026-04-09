-- Günlük ve haftalık toplam token tüketimini döner (UTC).
create or replace function public.llm_get_user_token_usage(p_user_id uuid)
returns table (daily_total integer, weekly_total integer)
language plpgsql
stable
security definer
set search_path = public
as $$
begin
    return query
    select
        coalesce(sum(case
            when l.assistant_received_at >= date_trunc('day', now() at time zone 'UTC')
            then coalesce(l.total_tokens, 0) else 0
        end), 0)::integer as daily_total,
        coalesce(sum(case
            when l.assistant_received_at >= (now() at time zone 'UTC') - interval '7 days'
            then coalesce(l.total_tokens, 0) else 0
        end), 0)::integer as weekly_total
    from public.llm_interaction_log l
    where l.user_id = p_user_id
      and l.outcome = 'ok';
end;
$$;

revoke all on function public.llm_get_user_token_usage(uuid) from public;
grant execute on function public.llm_get_user_token_usage(uuid) to service_role;
