-- Chat sessions table
create table if not exists public.chat_sessions (
    id         uuid primary key default gen_random_uuid(),
    user_id    uuid not null references auth.users(id) on delete cascade,
    title      text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- Chat messages table
create table if not exists public.chat_messages (
    id         uuid primary key default gen_random_uuid(),
    session_id uuid not null references public.chat_sessions(id) on delete cascade,
    user_id    uuid,
    role       text not null check (role in ('user', 'assistant', 'system')),
    content    text not null default '',
    created_at timestamptz not null default now()
);

-- Indexes
create index if not exists idx_chat_sessions_user_id on public.chat_sessions(user_id);
create index if not exists idx_chat_messages_session_id on public.chat_messages(session_id);
create index if not exists idx_chat_messages_created_at on public.chat_messages(session_id, created_at);

-- RLS
alter table public.chat_sessions enable row level security;
alter table public.chat_messages enable row level security;

-- Policies: users can only access their own sessions
create policy "Users can manage their own sessions"
    on public.chat_sessions
    for all
    using (auth.uid() = user_id)
    with check (auth.uid() = user_id);

-- Policies: users can manage messages in their own sessions
create policy "Users can manage messages in own sessions"
    on public.chat_messages
    for all
    using (
        exists (
            select 1 from public.chat_sessions cs
            where cs.id = chat_messages.session_id
            and cs.user_id = auth.uid()
        )
    )
    with check (
        exists (
            select 1 from public.chat_sessions cs
            where cs.id = chat_messages.session_id
            and cs.user_id = auth.uid()
        )
    );

-- updated_at trigger
create or replace function public.update_updated_at()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

create trigger chat_sessions_updated_at
    before update on public.chat_sessions
    for each row
    execute function public.update_updated_at();
