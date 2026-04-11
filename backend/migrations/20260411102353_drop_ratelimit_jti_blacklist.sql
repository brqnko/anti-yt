-- +goose Up
DROP TABLE t_ratelimit;
DROP TABLE t_jti_blacklist;

-- +goose Down
-- +goose StatementBegin
CREATE UNLOGGED TABLE t_jti_blacklist (
    jti uuid NOT NULL PRIMARY KEY,
    expires_at timestamptz NOT NULL
);

CREATE INDEX idx_1_t_jti_blacklist ON t_jti_blacklist (expires_at);

CREATE UNLOGGED TABLE t_ratelimit (
    t_ratelimit_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_public_id uuid NOT NULL,

    consumed_quota int NOT NULL DEFAULT 0,

    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE UNIQUE INDEX idx_1_t_ratelimit ON t_ratelimit (user_public_id);
-- +goose StatementEnd
