-- +goose Up

CREATE TABLE hits (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    path        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE hits;