-- İsteğe bağlı: hata sınıfı ve HTTP durumu; raporlama/filtreleme kolaylığı.
alter table public.llm_interaction_log
    add column if not exists error_code text,
    add column if not exists provider_http_status integer;

comment on column public.llm_interaction_log.error_code is
    'rate_limited, upstream_5xx, timeout, auth_failed, bad_request vb. (API yazar).';
comment on column public.llm_interaction_log.provider_http_status is
    'Provider HTTP durum kodu (429, 500, …); NULL ise bilinmiyor veya başarılı.';

create index if not exists idx_llm_interaction_log_error_code
    on public.llm_interaction_log (error_code)
    where error_code is not null;
