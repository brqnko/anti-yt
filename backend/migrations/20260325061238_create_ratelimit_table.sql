-- +goose Up
CREATE UNLOGGED TABLE t_ratelimit (
    t_ratelimit_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_public_id uuid NOT NULL, -- jwtからの検索を高速化するためuuidを使用

    consumed_quota int NOT NULL DEFAULT 0,

    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE UNIQUE INDEX idx_1_t_ratelimit ON t_ratelimit (user_public_id);

-- +goose Down
DROP TABLE t_ratelimit;
