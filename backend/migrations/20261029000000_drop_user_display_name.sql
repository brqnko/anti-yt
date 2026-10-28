-- +goose Up
ALTER TABLE m_user DROP COLUMN display_name;
ALTER TABLE h_user DROP COLUMN display_name;

-- +goose Down
-- Removed display names cannot be recovered; use the former registration default.
ALTER TABLE m_user ADD COLUMN display_name VARCHAR(32) NOT NULL DEFAULT 'anonymous';
ALTER TABLE h_user ADD COLUMN display_name VARCHAR(32) NOT NULL DEFAULT 'anonymous';
ALTER TABLE m_user ALTER COLUMN display_name DROP DEFAULT;
ALTER TABLE h_user ALTER COLUMN display_name DROP DEFAULT;
