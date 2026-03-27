-- name: GetVideoWatchTitlesByUser :many
SELECT
    vw.m_user_id AS user_id,
    u.language_code AS language_code,
    string_agg(DISTINCT REPLACE(video.external_title, ',', ''), ',') AS title_concat
FROM
    t_video_watch vw
INNER JOIN
    m_video video ON vw.m_video_id = video.m_video_id
INNER JOIN
    m_user u ON vw.m_user_id = u.m_user_id
WHERE
    @lower_id <= vw.public_id
GROUP BY
    vw.m_user_id, u.language_code;

-- name: UpsertMonthlyVideoWatchSummary :exec
INSERT INTO s_monthly_video_watch (
    m_user_id,
    ai_summary_title,
    ai_summary_description,
    ai_model,
    generated_at,
    target_month,
    public_id
) VALUES (
    @user_id,
    @ai_summary_title,
    @ai_summary_description,
    @ai_model,
    @generated_at,
    @target_month,
    uuidv7()
)
ON CONFLICT (m_user_id, target_month) DO UPDATE SET
    ai_summary_title = EXCLUDED.ai_summary_title,
    ai_summary_description = EXCLUDED.ai_summary_description,
    ai_model = EXCLUDED.ai_model,
    generated_at = EXCLUDED.generated_at,
    updated_at = current_timestamp;
