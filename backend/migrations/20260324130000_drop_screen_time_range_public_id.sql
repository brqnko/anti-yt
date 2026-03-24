-- +goose Up
DROP INDEX uk_1_m_user_screen_time_range;
ALTER TABLE m_user_screen_time_range DROP COLUMN public_id;

-- +goose Down
ALTER TABLE m_user_screen_time_range ADD COLUMN public_id uuid NOT NULL DEFAULT uuidv7();
CREATE UNIQUE INDEX uk_1_m_user_screen_time_range ON m_user_screen_time_range (public_id);
