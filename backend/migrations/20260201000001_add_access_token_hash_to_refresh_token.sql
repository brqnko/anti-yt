-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_refresh_token
    ADD COLUMN access_token_hash VARCHAR(256) NOT NULL DEFAULT '';
ALTER TABLE m_refresh_token
    ALTER COLUMN access_token_hash DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_refresh_token
    DROP COLUMN access_token_hash;
-- +goose StatementEnd
