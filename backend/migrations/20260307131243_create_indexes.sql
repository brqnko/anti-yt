-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_1_m_user ON m_user (m_user_authorization_id);
CREATE INDEX idx_2_m_refresh_token ON m_refresh_token (m_user_authorization_id, updated_at);

ALTER TABLE m_refresh_token DROP COLUMN access_token_hash;

DELETE FROM m_refresh_token;
ALTER TABLE m_refresh_token ADD COLUMN access_token_jti UUID NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE m_refresh_token DROP COLUMN access_token_jti;

ALTER TABLE m_refresh_token
    ADD COLUMN access_token_hash VARCHAR(256) NOT NULL DEFAULT '';
ALTER TABLE m_refresh_token
    ALTER COLUMN access_token_hash DROP DEFAULT;

DROP INDEX idx_2_m_refresh_token;
DROP INDEX idx_1_m_user;
-- +goose StatementEnd
