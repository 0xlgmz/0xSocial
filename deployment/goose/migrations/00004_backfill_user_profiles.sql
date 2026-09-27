-- +goose Up

-- Registration creates profiles for new users. Backfill users created before
-- that behavior was introduced, while preserving any existing profiles.
INSERT INTO profiles (user_id)
SELECT id
FROM users
ON CONFLICT (user_id) DO NOTHING;

-- +goose Down

-- Intentionally left blank. A rollback must not delete user profile data.
