-- +goose Up
-- +goose StatementBegin
CREATE TABLE albums (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    parent_id VARCHAR(255) REFERENCES albums(id) ON DELETE CASCADE
);

CREATE TABLE media (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL,
	captured_at TIMESTAMP DEFAULT (now() AT TIME ZONE 'UTC') NOT NULL,
    album_id VARCHAR(255) NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    hash VARCHAR(255) NOT NULL,
    thumbnail BYTEA,
	exif JSONB NOT NULL,
    media_type VARCHAR(10) NOT NULL
);

ALTER TABLE albums
	ADD COLUMN thumbnail_id VARCHAR(255) REFERENCES media(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE albums DROP COLUMN thumbnail_id;
DROP TABLE media;
DROP TABLE albums;
-- +goose StatementEnd
