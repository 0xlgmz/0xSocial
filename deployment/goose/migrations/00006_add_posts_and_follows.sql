-- +goose Up

CREATE TABLE follows (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    follower_id  BIGINT NOT NULL
                 REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT NOT NULL
                 REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT follows_distinct_users_check
        CHECK (follower_id <> following_id),
    CONSTRAINT follows_follower_following_key
        UNIQUE (follower_id, following_id)
);

CREATE INDEX follows_follower_id_id_idx
    ON follows (follower_id, id DESC);

CREATE INDEX follows_following_id_id_idx
    ON follows (following_id, id DESC);

CREATE TABLE posts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL
               REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT posts_content_length_check
        CHECK (CHAR_LENGTH(BTRIM(content)) BETWEEN 1 AND 2000),
    CONSTRAINT posts_updated_at_check
        CHECK (updated_at >= created_at),
    CONSTRAINT posts_deleted_at_check
        CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

CREATE INDEX posts_active_user_id_id_idx
    ON posts (user_id, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX posts_active_id_idx
    ON posts (id DESC)
    WHERE deleted_at IS NULL;

-- +goose Down

DROP TABLE posts;
DROP TABLE follows;
