-- +goose Up
CREATE INDEX idx_2_m_user_subscribing_channel ON m_user_subscribing_channel (m_user_id, public_id);

-- +goose Down
DROP INDEX idx_2_m_user_subscribing_channel;
