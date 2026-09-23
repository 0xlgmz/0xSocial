-- +goose Up

-- Authentication data is deliberately split across tables so additional
-- authentication methods (passkeys, OAuth, MFA) can be added without changing
-- the core user record.

CREATE TABLE users (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email                TEXT NOT NULL,
    email_normalized     TEXT NOT NULL,
    email_verified_at    TIMESTAMPTZ,
    status               TEXT NOT NULL DEFAULT 'pending_verification',
    sessions_valid_after TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_length_check
        CHECK (CHAR_LENGTH(email) BETWEEN 3 AND 320),
    CONSTRAINT users_email_normalized_check
        CHECK (
            CHAR_LENGTH(email_normalized) BETWEEN 3 AND 320
            AND email_normalized = LOWER(BTRIM(email_normalized))
        ),
    CONSTRAINT users_status_check
        CHECK (status IN ('pending_verification', 'active', 'locked', 'disabled')),
    CONSTRAINT users_email_normalized_key UNIQUE (email_normalized)
);

CREATE TABLE password_credentials (
    user_id             BIGINT PRIMARY KEY
                        REFERENCES users(id) ON DELETE CASCADE,
    password_hash       TEXT NOT NULL,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT password_credentials_hash_check
        CHECK (CHAR_LENGTH(password_hash) > 0)
);

-- Store only SHA-256(token) in token_hash. The raw random token belongs only
-- in the Secure, HttpOnly session cookie.
CREATE TABLE sessions (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        BIGINT NOT NULL
                   REFERENCES users(id) ON DELETE CASCADE,
    token_hash     BYTEA NOT NULL,
    device_id      TEXT,
    ip_address     INET,
    user_agent     TEXT,
    auth_method    TEXT NOT NULL DEFAULT 'password',
    risk_level     TEXT NOT NULL DEFAULT 'normal',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ,
    revoked_reason TEXT,

    CONSTRAINT sessions_token_hash_key UNIQUE (token_hash),
    CONSTRAINT sessions_token_hash_length_check
        CHECK (OCTET_LENGTH(token_hash) = 32),
    CONSTRAINT sessions_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT sessions_last_seen_check
        CHECK (last_seen_at >= created_at),
    CONSTRAINT sessions_risk_level_check
        CHECK (risk_level IN ('normal', 'elevated', 'high')),
    CONSTRAINT sessions_revocation_check
        CHECK (
            (revoked_at IS NULL AND revoked_reason IS NULL)
            OR (revoked_at IS NOT NULL AND revoked_reason IS NOT NULL)
        )
);

CREATE INDEX sessions_user_id_idx
    ON sessions (user_id);

CREATE INDEX sessions_active_user_expiry_idx
    ON sessions (user_id, expires_at)
    WHERE revoked_at IS NULL;

CREATE INDEX sessions_expires_at_idx
    ON sessions (expires_at);

-- Used for both email verification and password recovery. Store only the
-- SHA-256 digest of the token sent by email, never the raw token.
CREATE TABLE auth_tokens (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL
               REFERENCES users(id) ON DELETE CASCADE,
    purpose    TEXT NOT NULL,
    token_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,

    CONSTRAINT auth_tokens_token_hash_key UNIQUE (token_hash),
    CONSTRAINT auth_tokens_token_hash_length_check
        CHECK (OCTET_LENGTH(token_hash) = 32),
    CONSTRAINT auth_tokens_purpose_check
        CHECK (purpose IN ('verify_email', 'reset_password')),
    CONSTRAINT auth_tokens_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT auth_tokens_used_at_check
        CHECK (used_at IS NULL OR used_at >= created_at)
);

CREATE INDEX auth_tokens_user_id_idx
    ON auth_tokens (user_id);

CREATE INDEX auth_tokens_active_user_purpose_idx
    ON auth_tokens (user_id, purpose, expires_at)
    WHERE used_at IS NULL;

-- Authentication events form an append-only audit stream and can later feed
-- fraud detection. Passwords, raw tokens, and full request bodies must never be
-- written to this table or to metadata.
CREATE TABLE auth_events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    session_id  BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    event_type  TEXT NOT NULL,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT auth_events_event_type_check
        CHECK (CHAR_LENGTH(BTRIM(event_type)) > 0),
    CONSTRAINT auth_events_metadata_object_check
        CHECK (JSONB_TYPEOF(metadata) = 'object')
);

CREATE INDEX auth_events_user_created_at_idx
    ON auth_events (user_id, created_at DESC);

CREATE INDEX auth_events_session_created_at_idx
    ON auth_events (session_id, created_at DESC);

CREATE INDEX auth_events_type_created_at_idx
    ON auth_events (event_type, created_at DESC);

-- +goose Down

DROP TABLE auth_events;
DROP TABLE auth_tokens;
DROP TABLE sessions;
DROP TABLE password_credentials;
DROP TABLE users;
