-- +goose Up
-- +goose StatementBegin
CREATE TABLE t_video_watched
(
    t_video_watched_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id          bigint NOT NULL,
    m_video_id         bigint NOT NULL
);

CREATE UNIQUE INDEX uk_1_t_video_watched ON t_video_watched (m_user_id, m_video_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE t_video_watched;
-- +goose StatementEnd
