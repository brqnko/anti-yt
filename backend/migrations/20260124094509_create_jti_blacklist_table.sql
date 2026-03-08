-- +goose Up
-- +goose StatementBegin
CREATE UNLOGGED TABLE t_jti_blacklist (
    jti uuid NOT NULL PRIMARY KEY,
    expires_at timestamptz NOT NULL
);

CREATE INDEX idx_1_t_jti_blacklist ON t_jti_blacklist (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE t_jti_blacklist;
-- +goose StatementEnd
