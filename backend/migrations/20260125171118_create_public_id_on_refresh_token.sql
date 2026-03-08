-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_refresh_token
    ADD COLUMN public_id uuid NOT NULL DEFAULT uuidv7();
CREATE UNIQUE INDEX uk_2_m_refresh_token ON m_refresh_token (public_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_refresh_token
    DROP COLUMN public_id;
-- +goose StatementEnd
