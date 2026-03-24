-- +goose Up
CREATE TABLE m_valuable_channel (
    m_valuable_channel_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_channel_id bigint NOT NULL,

    category_code int NOT NULL,
    valuable_description VARCHAR(256) NOT NULL,

    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE INDEX uk_1_m_valuable_channel ON m_valuable_channel (m_channel_id);

-- +goose Down
DROP TABLE m_valuable_channel;
