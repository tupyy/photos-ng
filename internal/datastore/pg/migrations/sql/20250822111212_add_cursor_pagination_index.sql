-- +goose Up
-- +goose StatementBegin
-- Add composite index for cursor-based pagination on media table
-- This index supports efficient queries with WHERE captured_at < ? OR (captured_at = ? AND id < ?)
-- and ORDER BY captured_at DESC, id DESC
CREATE INDEX idx_media_cursor_pagination ON media(captured_at DESC, id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_media_cursor_pagination;
-- +goose StatementEnd