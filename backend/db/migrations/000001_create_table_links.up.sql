CREATE TABLE links (
    id           BIGSERIAL PRIMARY KEY,
    code         VARCHAR(16) NOT NULL,
    original_url TEXT        NOT NULL,
    clicks       BIGINT      NOT NULL DEFAULT 0,
    expires_at   TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT links_code_key UNIQUE (code)
);