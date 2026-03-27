-- +goose Up
ALTER TABLE s_monthly_video_watch DROP COLUMN w_monthly_video_watch_id;
DROP TABLE w_monthly_video_watch;

-- +goose Down
CREATE TABLE w_monthly_video_watch
(
    w_monthly_video_watch_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    m_user_id                bigint       NOT NULL,
    batch_status_code        int          NOT NULL,
    ai_model                 VARCHAR(128) NOT NULL,
    started_at               timestamptz  NOT NULL,
    finished_at              timestamptz  NOT NULL,
    target_month             date         NOT NULL,
    fail_reason              VARCHAR(128) NOT NULL,

    created_at               timestamptz  NOT NULL DEFAULT current_timestamp,
    updated_at               timestamptz  NOT NULL DEFAULT current_timestamp
);

ALTER TABLE s_monthly_video_watch ADD COLUMN w_monthly_video_watch_id bigint NOT NULL DEFAULT 0;
