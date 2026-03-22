-- 認証テーブルにIssuerとSubを保存する
-- IssuerとSubの組み合わせで一意制約があり、既に存在する場合は、何もしない。
-- 重複した場合でも特にエラーは発生せず、既存のレコードが返される。
-- 返り値は、m_user_authorization_idとpublic_id、そして新規作成されたかどうかを示すis_createdフラグ。
-- name: SaveAuthorization :one
INSERT INTO
    m_user_authorization (issuer, sub, last_logged_in_at, public_id)
VALUES
    ($1, $2, $3, $4)
ON CONFLICT (issuer, sub) DO
UPDATE
SET
    issuer = EXCLUDED.issuer,
    last_logged_in_at = EXCLUDED.last_logged_in_at,
    updated_at = current_timestamp
RETURNING
    m_user_authorization_id,
    (xmax = 0) AS is_created;

-- authorization_idから、そのユーザーのpublic_idを取得する。
-- authorization_idが存在しない場合はpgx.ErrNoRowsが返される。
-- 退会済みの場合はtrueが返ってくる。
-- name: GetUserIDByAuthorization :one
SELECT
    m_user.public_id,
    false AS is_deactivated
FROM
    m_user
WHERE
    m_user.m_user_authorization_id = $1
UNION ALL
SELECT
    h_user.public_id,
    true AS is_deactivated
FROM
    h_user
WHERE
    h_user.m_user_authorization_id = $1
ORDER BY
    is_deactivated
LIMIT
    1;
