-- m_user.public_idから、そのユーザーが今日視聴していた合計時間(seconds)を返す。
-- その日に一本も動画を視聴していない場合は0を返します。
-- name: GetTotalWatchSeconds :one
SELECT COALESCE(EXTRACT(EPOCH FROM SUM(t_video_watch.watch_end_at - t_video_watch.watch_start_at)), 0)::int FROM t_video_watch
WHERE t_video_watch.m_user_id = (SELECT m_user.m_user_id FROM m_user WHERE m_user.public_id = @public_id)
    AND CURRENT_DATE <= t_video_watch.watch_start_at AND t_video_watch.watch_start_at < CURRENT_DATE + INTERVAL '1 day';