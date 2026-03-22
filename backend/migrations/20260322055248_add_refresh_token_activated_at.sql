-- +goose Up
ALTER TABLE m_refresh_token ADD COLUMN activated_at timestamptz DEFAULT current_timestamp NOT NULL;
ALTER TABLE m_refresh_token ALTER COLUMN activated_at DROP DEFAULT;
ALTER TABLE m_refresh_token ADD COLUMN last_logged_in_at timestamptz DEFAULT current_timestamp NOT NULL;
ALTER TABLE m_refresh_token ALTER COLUMN last_logged_in_at DROP DEFAULT;

-- +goose Down
ALTER TABLE m_refresh_token DROP COLUMN last_logged_in_at;
ALTER TABLE m_refresh_token DROP COLUMN activated_at;
