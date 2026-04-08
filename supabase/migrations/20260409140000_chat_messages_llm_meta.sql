-- Asistan (ve istenirse ileride user) satırı başına hangi LLM ile üretildiği
alter table public.chat_messages
    add column if not exists provider text,
    add column if not exists model text;

comment on column public.chat_messages.provider is 'Bu mesaj asistan ise kullanılan provider (openai, gemini, …)';
comment on column public.chat_messages.model is 'Bu mesaj asistan ise kullanılan model id';
