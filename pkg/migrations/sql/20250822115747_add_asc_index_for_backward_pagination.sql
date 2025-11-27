-- +goose Up
-- +goose StatementBegin
-- Add ascending index for backward cursor pagination
-- This index supports efficient backward navigation with ORDER BY captured_at ASC, id ASC
-- Used when navigating backward in timeline (toward newer photos)
CREATE INDEX idx_media_cursor_pagination_asc ON media(captured_at ASC, id ASC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_cursor_pagination_asc;
-- +goose StatementEnd