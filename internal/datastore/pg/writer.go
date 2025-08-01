package pg

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Writer struct {
	tx pgx.Tx
}

// WriteAlbum creates or updates an album using PostgreSQL upsert (ON CONFLICT)
func (w *Writer) WriteAlbum(ctx context.Context, album entity.Album) error {
	// Build the upsert statement
	stmt := psql.Insert(albumsTable).
		Columns(
			albumID,
			albumCreatedAt,
			albumPath,
			albumDescription,
			albumParentID,
			albumThumbnailID,
		).
		Values(
			album.ID,
			album.CreatedAt,
			album.Path,
			album.Description,
			album.Parent,
			album.Thumbnail,
		).
		Suffix("ON CONFLICT (" + albumID + ") DO UPDATE SET " +
			albumDescription + " = EXCLUDED." + albumDescription + ", " +
			albumThumbnailID + " = EXCLUDED." + albumThumbnailID)

	// Convert to SQL
	sql, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	// Execute the query
	_, err = w.tx.Exec(ctx, sql, args...)
	return err
}

// DeleteAlbum deletes an album from the database
func (w *Writer) DeleteAlbum(ctx context.Context, albumID string) error {
	// Build the delete statement
	stmt := psql.Delete(albumsTable).
		Where(sq.Eq{albumID: albumID})

	// Convert to SQL
	sql, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	// Execute the query
	_, err = w.tx.Exec(ctx, sql, args...)
	return err
}
