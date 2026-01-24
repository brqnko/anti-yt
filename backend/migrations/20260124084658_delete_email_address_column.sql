-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_user_authorization DROP COLUMN email_address;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_user_authorization ADD COLUMN email_address VARCHAR(256) NOT NULL DEFAULT '';
-- +goose StatementEnd
