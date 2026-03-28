-- +goose Up
-- +goose StatementBegin
INSERT INTO m_user_authorization (m_user_authorization_id, issuer, sub, last_logged_in_at, public_id)
OVERRIDING SYSTEM VALUE
VALUES (0, 'system', 'system', CURRENT_TIMESTAMP, '00000000-0000-0000-0000-000000000001');

INSERT INTO m_user (m_user_id, m_user_authorization_id, display_name, language_code, daily_screen_time_seconds, joined_at, public_id)
OVERRIDING SYSTEM VALUE
VALUES (0, 0, 'official', 'ja', 86401, CURRENT_TIMESTAMP, '00000000-0000-0000-0000-000000000000');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM m_user WHERE m_user_id = 0;
DELETE FROM m_user_authorization WHERE m_user_authorization_id = 0;
-- +goose StatementEnd
