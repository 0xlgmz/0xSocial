-- +goose Up
CREATE TABLE profiles (
    user_id      BIGINT PRIMARY KEY
                 REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL DEFAULT '',
    bio          TEXT NOT NULL DEFAULT '',
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT profiles_display_name_length_check
        CHECK (CHAR_LENGTH(display_name) <= 80),
    CONSTRAINT profiles_bio_length_check
        CHECK (CHAR_LENGTH(bio) <= 500)
);

-- +goose Down
DROP TABLE profiles;
