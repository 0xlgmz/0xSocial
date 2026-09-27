-- +goose Up

CREATE TABLE post_media (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT NOT NULL
                REFERENCES users(id) ON DELETE CASCADE,
    post_id     BIGINT
                REFERENCES posts(id) ON DELETE CASCADE,
    object_key  TEXT NOT NULL,
    mime_type   TEXT NOT NULL,
    size_bytes  BIGINT NOT NULL,
    position    SMALLINT,
    status      TEXT NOT NULL DEFAULT 'pending',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attached_at TIMESTAMPTZ,

    CONSTRAINT post_media_object_key_key UNIQUE (object_key),
    CONSTRAINT post_media_post_position_key UNIQUE (post_id, position),
    CONSTRAINT post_media_object_key_check
        CHECK (CHAR_LENGTH(BTRIM(object_key)) BETWEEN 1 AND 1024),
    CONSTRAINT post_media_mime_type_check
        CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/webp')),
    CONSTRAINT post_media_size_check
        CHECK (size_bytes BETWEEN 1 AND 10485760),
    CONSTRAINT post_media_position_check
        CHECK (position IS NULL OR position BETWEEN 0 AND 3),
    CONSTRAINT post_media_status_check
        CHECK (status IN ('pending', 'attached', 'deleted')),
    CONSTRAINT post_media_expiry_check
        CHECK (expires_at > created_at),
    CONSTRAINT post_media_state_check
        CHECK (
            (
                status = 'pending'
                AND post_id IS NULL
                AND position IS NULL
                AND attached_at IS NULL
            )
            OR (
                status IN ('attached', 'deleted')
                AND post_id IS NOT NULL
                AND position IS NOT NULL
                AND attached_at IS NOT NULL
            )
        )
);

CREATE INDEX post_media_pending_expiry_idx
    ON post_media (expires_at)
    WHERE status = 'pending';

CREATE INDEX post_media_attached_post_idx
    ON post_media (post_id, position)
    WHERE status = 'attached';

-- +goose Down

DROP TABLE post_media;
