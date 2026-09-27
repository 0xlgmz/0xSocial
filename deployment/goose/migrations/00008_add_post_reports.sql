-- +goose Up

CREATE TABLE post_reports (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    post_id      BIGINT NOT NULL
                 REFERENCES posts(id) ON DELETE RESTRICT,
    reporter_id  BIGINT NOT NULL
                 REFERENCES users(id) ON DELETE RESTRICT,
    reason       TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    reviewed_by  BIGINT
                 REFERENCES users(id) ON DELETE SET NULL,
    review_notes TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at  TIMESTAMPTZ,

    CONSTRAINT post_reports_post_reporter_key
        UNIQUE (post_id, reporter_id),
    CONSTRAINT post_reports_reason_check
        CHECK (reason IN (
            'spam',
            'harassment',
            'hate_speech',
            'violence',
            'sexual_content',
            'self_harm',
            'false_information',
            'other'
        )),
    CONSTRAINT post_reports_status_check
        CHECK (status IN ('pending', 'reviewing', 'actioned', 'dismissed')),
    CONSTRAINT post_reports_review_state_check
        CHECK (
            (status IN ('pending', 'reviewing') AND reviewed_at IS NULL)
            OR
            (status IN ('actioned', 'dismissed') AND reviewed_at IS NOT NULL)
        ),
    CONSTRAINT post_reports_updated_at_check
        CHECK (updated_at >= created_at),
    CONSTRAINT post_reports_reviewed_at_check
        CHECK (reviewed_at IS NULL OR reviewed_at >= created_at)
);

CREATE INDEX post_reports_moderation_queue_idx
    ON post_reports (status, created_at, id)
    WHERE status IN ('pending', 'reviewing');

CREATE INDEX post_reports_post_id_idx
    ON post_reports (post_id, id DESC);

-- +goose Down

DROP TABLE post_reports;
