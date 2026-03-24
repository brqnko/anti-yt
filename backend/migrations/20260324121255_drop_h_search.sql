-- +goose Up
DROP TABLE h_search;

-- +goose Down
CREATE TABLE h_search
(
    h_search_id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id      bigint       NOT NULL,
    search_keyword VARCHAR(256) NOT NULL,
    searched_at    timestamptz  NOT NULL DEFAULT current_timestamp,

    created_at     timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at     timestamptz  NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_1_h_search ON h_search (m_user_id);

