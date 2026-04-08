-- Son başarılı LLM cevabında kullanılan provider/model (GET /chats/:id için)
alter table public.chat_sessions
    add column if not exists last_provider text,
    add column if not exists last_model text;

comment on column public.chat_sessions.last_provider is 'Son asistan yanıtında kullanılan provider (örn. openai, gemini)';
comment on column public.chat_sessions.last_model is 'Son asistan yanıtında kullanılan model id';
