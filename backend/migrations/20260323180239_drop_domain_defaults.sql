-- +goose Up
-- +goose StatementBegin

-- m_user
ALTER TABLE m_user ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_user ALTER COLUMN joined_at DROP DEFAULT;
ALTER TABLE m_user ALTER COLUMN daily_screen_time_seconds DROP DEFAULT;

-- m_user_authorization
ALTER TABLE m_user_authorization ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_user_authorization ALTER COLUMN last_logged_in_at DROP DEFAULT;

-- m_refresh_token
ALTER TABLE m_refresh_token ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_refresh_token ALTER COLUMN generation DROP DEFAULT;

-- h_user
ALTER TABLE h_user ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE h_user ALTER COLUMN left_at DROP DEFAULT;

-- m_user_screen_time_range
ALTER TABLE m_user_screen_time_range ALTER COLUMN public_id DROP DEFAULT;

-- m_channel
ALTER TABLE m_channel ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_channel ALTER COLUMN fetched_at DROP DEFAULT;
ALTER TABLE m_channel ALTER COLUMN rss_fetched_at DROP DEFAULT;

-- m_video
ALTER TABLE m_video ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_video ALTER COLUMN fetched_at DROP DEFAULT;
ALTER TABLE m_video ALTER COLUMN external_created_at DROP DEFAULT;
ALTER TABLE m_video ALTER COLUMN external_thumbnail_url DROP DEFAULT;

-- t_video_watch
ALTER TABLE t_video_watch ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE t_video_watch ALTER COLUMN watch_start_at DROP DEFAULT;
ALTER TABLE t_video_watch ALTER COLUMN watch_end_at DROP DEFAULT;
ALTER TABLE t_video_watch ALTER COLUMN watch_position_seconds DROP DEFAULT;

-- h_search
ALTER TABLE h_search ALTER COLUMN searched_at DROP DEFAULT;

-- s_monthly_video_watch
ALTER TABLE s_monthly_video_watch ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE s_monthly_video_watch ALTER COLUMN generated_at DROP DEFAULT;

-- w_monthly_video_watch
ALTER TABLE w_monthly_video_watch ALTER COLUMN batch_status_code DROP DEFAULT;
ALTER TABLE w_monthly_video_watch ALTER COLUMN started_at DROP DEFAULT;
ALTER TABLE w_monthly_video_watch ALTER COLUMN finished_at DROP DEFAULT;
ALTER TABLE w_monthly_video_watch ALTER COLUMN target_month DROP DEFAULT;
ALTER TABLE w_monthly_video_watch ALTER COLUMN fail_reason DROP DEFAULT;

-- m_playlist
ALTER TABLE m_playlist ALTER COLUMN public_id DROP DEFAULT;
ALTER TABLE m_playlist ALTER COLUMN visibility_code DROP DEFAULT;
ALTER TABLE m_playlist ALTER COLUMN playlist_code DROP DEFAULT;
ALTER TABLE m_playlist ALTER COLUMN video_count DROP DEFAULT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- m_playlist
ALTER TABLE m_playlist ALTER COLUMN video_count SET DEFAULT 0;
ALTER TABLE m_playlist ALTER COLUMN playlist_code SET DEFAULT 0;
ALTER TABLE m_playlist ALTER COLUMN visibility_code SET DEFAULT 0;
ALTER TABLE m_playlist ALTER COLUMN public_id SET DEFAULT uuidv7();

-- w_monthly_video_watch
ALTER TABLE w_monthly_video_watch ALTER COLUMN fail_reason SET DEFAULT '';
ALTER TABLE w_monthly_video_watch ALTER COLUMN target_month SET DEFAULT '1970-01-01';
ALTER TABLE w_monthly_video_watch ALTER COLUMN finished_at SET DEFAULT '9999-12-31 23:59:59UTC';
ALTER TABLE w_monthly_video_watch ALTER COLUMN started_at SET DEFAULT current_timestamp;
ALTER TABLE w_monthly_video_watch ALTER COLUMN batch_status_code SET DEFAULT 1;

-- s_monthly_video_watch
ALTER TABLE s_monthly_video_watch ALTER COLUMN generated_at SET DEFAULT current_timestamp;
ALTER TABLE s_monthly_video_watch ALTER COLUMN public_id SET DEFAULT uuidv7();

-- h_search
ALTER TABLE h_search ALTER COLUMN searched_at SET DEFAULT current_timestamp;

-- t_video_watch
ALTER TABLE t_video_watch ALTER COLUMN watch_position_seconds SET DEFAULT 0;
ALTER TABLE t_video_watch ALTER COLUMN watch_end_at SET DEFAULT '9999-12-31 23:59:59UTC';
ALTER TABLE t_video_watch ALTER COLUMN watch_start_at SET DEFAULT current_timestamp;
ALTER TABLE t_video_watch ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_video
ALTER TABLE m_video ALTER COLUMN external_thumbnail_url SET DEFAULT '';
ALTER TABLE m_video ALTER COLUMN external_created_at SET DEFAULT '1970-01-01 00:00:00UTC';
ALTER TABLE m_video ALTER COLUMN fetched_at SET DEFAULT current_timestamp;
ALTER TABLE m_video ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_channel
ALTER TABLE m_channel ALTER COLUMN rss_fetched_at SET DEFAULT current_timestamp;
ALTER TABLE m_channel ALTER COLUMN fetched_at SET DEFAULT current_timestamp;
ALTER TABLE m_channel ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_user_screen_time_range
ALTER TABLE m_user_screen_time_range ALTER COLUMN public_id SET DEFAULT uuidv7();

-- h_user
ALTER TABLE h_user ALTER COLUMN left_at SET DEFAULT current_timestamp;
ALTER TABLE h_user ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_refresh_token
ALTER TABLE m_refresh_token ALTER COLUMN generation SET DEFAULT 1;
ALTER TABLE m_refresh_token ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_user_authorization
ALTER TABLE m_user_authorization ALTER COLUMN last_logged_in_at SET DEFAULT current_timestamp;
ALTER TABLE m_user_authorization ALTER COLUMN public_id SET DEFAULT uuidv7();

-- m_user
ALTER TABLE m_user ALTER COLUMN daily_screen_time_seconds SET DEFAULT 86401;
ALTER TABLE m_user ALTER COLUMN joined_at SET DEFAULT current_timestamp;
ALTER TABLE m_user ALTER COLUMN public_id SET DEFAULT uuidv7();

-- +goose StatementEnd
