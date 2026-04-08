-- Oturum açılırken hedeflenen provider/model (henüz mesaj yokken GET özetinde kullanılır)
alter table public.chat_sessions
    add column if not exists default_provider text,
    add column if not exists default_model text;

comment on column public.chat_sessions.default_provider is 'Oturum oluşturulurken seçilen veya API default provider';
comment on column public.chat_sessions.default_model is 'Oturum oluşturulurken seçilen veya API default model';
