CREATE TABLE public.email_verification_codes
(
    id            character varying(64)                  NOT NULL,
    user_id       character varying(64)                  NOT NULL,
    email         character varying(256)                 NOT NULL,
    purpose       character varying(64)                  NOT NULL,
    code_hash     character varying(255)                 NOT NULL,
    status        integer                  DEFAULT 0     NOT NULL,
    attempt_count integer                  DEFAULT 0     NOT NULL,
    max_attempts  integer                  DEFAULT 5     NOT NULL,
    expires_at    timestamp with time zone               NOT NULL,
    sent_at       timestamp with time zone               NOT NULL,
    used_at       timestamp with time zone,
    created_at    timestamp with time zone DEFAULT now() NOT NULL,
    updated_at    timestamp with time zone DEFAULT now() NOT NULL,
    created_by    character varying(64)                  NOT NULL,
    updated_by    character varying(64)                  NOT NULL
);

ALTER TABLE ONLY public.email_verification_codes
    ADD CONSTRAINT email_verification_codes_pk PRIMARY KEY (id);

ALTER TABLE ONLY public.email_verification_codes
    ADD CONSTRAINT email_verification_codes_user_fk FOREIGN KEY (user_id) REFERENCES public.users (id);

CREATE INDEX email_verification_codes_idx_user_purpose_status
    ON public.email_verification_codes (user_id, purpose, status);

CREATE INDEX email_verification_codes_idx_email_purpose_status
    ON public.email_verification_codes (email, purpose, status);

CREATE INDEX email_verification_codes_idx_expires_at
    ON public.email_verification_codes (expires_at);

CREATE UNIQUE INDEX email_verification_codes_uq_active_user_purpose
    ON public.email_verification_codes (user_id, purpose)
    WHERE status = 0;

CREATE UNIQUE INDEX users_uq_email
    ON public.users (email)
    WHERE email IS NOT NULL AND email <> '';

