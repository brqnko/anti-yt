-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS m_comment;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE m_comment
(
    m_comment_id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    m_video_id                 bigint        NOT NULL,
    external_id                VARCHAR(512)  NOT NULL,
    external_content           VARCHAR(8192) NOT NULL,
    external_like_count        bigint        NOT NULL,
    external_user_id           VARCHAR(32)   NOT NULL,
    external_user_display_name VARCHAR(64)   NOT NULL,
    external_user_custom_id    VARCHAR(64)   NOT NULL,
    external_edited_flg        BOOLEAN       NOT NULL,
    external_created_at        timestamptz   NOT NULL,
    fetched_at                 timestamptz   NOT NULL DEFAULT current_timestamp,
    created_at                 timestamptz   NOT NULL DEFAULT current_timestamp,
    updated_at                 timestamptz   NOT NULL DEFAULT current_timestamp,
    public_id                  uuid          NOT NULL DEFAULT uuidv7()
);

CREATE INDEX idx_1_m_comment ON m_comment (m_video_id);
CREATE UNIQUE INDEX uk_1_m_comment ON m_comment (public_id);
CREATE UNIQUE INDEX uk_2_m_comment ON m_comment (external_id);
-- +goose StatementEnd
