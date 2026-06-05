-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_refresh_token DROP COLUMN device_fingerprint;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_refresh_token ADD COLUMN device_fingerprint VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE m_refresh_token ALTER COLUMN device_fingerprint DROP DEFAULT;
-- +goose StatementEnd
