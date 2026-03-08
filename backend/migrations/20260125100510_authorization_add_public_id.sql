-- +goose Up
-- +goose StatementBegin
ALTER TABLE m_user_authorization ADD COLUMN public_id uuid NOT NULL DEFAULT uuidv7();
CREATE UNIQUE INDEX uk_2_m_user_authorization ON m_user_authorization (public_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_user_authorization DROP COLUMN public_id;
-- +goose StatementEnd
