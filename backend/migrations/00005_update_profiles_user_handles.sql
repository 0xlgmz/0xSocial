-- +goose Up

ALTER TABLE profiles
ADD COLUMN handle TEXT;

-- Existing users need a unique value before the column becomes NOT NULL.
UPDATE profiles
SET handle = 'user_' || user_id;

ALTER TABLE profiles
ALTER COLUMN handle SET NOT NULL;

ALTER TABLE profiles
ADD CONSTRAINT profiles_handle_format_check
CHECK (handle ~ '^[a-z][a-z0-9_-]{2,29}$');

ALTER TABLE profiles
ADD CONSTRAINT profiles_handle_key UNIQUE (handle);

-- +goose Down

ALTER TABLE profiles
DROP COLUMN handle;
