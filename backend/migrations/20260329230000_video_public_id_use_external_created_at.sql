-- +goose Up
-- +goose StatementBegin

-- public_idのタイムスタンプビット(先頭48bit)をexternal_created_atに置き換える
UPDATE m_video SET public_id = (
    lpad(to_hex((extract(epoch from external_created_at) * 1000)::bigint), 12, '0')
    || substring(replace(public_id::text, '-', '') from 13)
)::uuid;

-- public_idにexternal_created_atが埋め込まれたので、複合ソートキーが不要になった
DROP INDEX idx_3_m_video;

-- public_id単独でソートできるようになったので、(m_channel_id, public_id DESC) で十分
CREATE INDEX idx_2_m_video ON m_video (m_channel_id, public_id DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX idx_2_m_video;
CREATE INDEX idx_3_m_video ON m_video (m_channel_id, external_created_at DESC, public_id DESC);

-- タイムスタンプビットの復元は不可能(元のデータが失われている)

-- +goose StatementEnd
