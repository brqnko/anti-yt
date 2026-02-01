-- name: CreateUser :one
INSERT INTO m_user (m_user_authorization_id, display_name, language_code, daily_screen_time_seconds)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
RETURNING m_user_id, display_name, language_code, joined_at, daily_screen_time_seconds, public_id;

-- name: CreateUserScreenTimeRanges :copyfrom
INSERT INTO m_user_screen_time_range (m_user_id, screen_time_range_start, screen_time_range_end)
VALUES ($1, $2, $3);

-- name: GetUserScreenTimeRanges :many
SELECT m_user_id, public_id, screen_time_range_start, screen_time_range_end FROM m_user_screen_time_range;
