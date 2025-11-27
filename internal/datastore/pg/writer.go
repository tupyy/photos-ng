package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

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
			album.ParentId,
			album.Thumbnail,
		).
		Suffix("ON CONFLICT ( id ) DO UPDATE SET " +
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

// WriteMedia creates or updates a media item using PostgreSQL upsert (ON CONFLICT)
func (w *Writer) WriteMedia(ctx context.Context, media entity.Media) error {
	// Convert EXIF metadata to JSON
	exifData, err := json.Marshal(media.Exif)
	if err != nil {
		return err
	}

	// Build the upsert statement
	stmt := psql.Insert(mediaTable).
		Columns(
			mediaID,
			mediaCreatedAt,
			mediaCapturedAt,
			mediaAlbumID,
			mediaFileName,
			mediaThumbnail,
			mediaExif,
			mediaMediaType,
			mediaHash,
		).
		Values(
			media.ID,
			time.Now(), // Use current time for created_at
			media.CapturedAt,
			media.Album.ID,
			media.Filename,
			media.Thumbnail,
			exifData,
			string(media.MediaType),
			media.Hash,
		).
		Suffix("ON CONFLICT (" + mediaID + ") DO UPDATE SET " +
			mediaCapturedAt + " = EXCLUDED." + mediaCapturedAt + ", " +
			mediaThumbnail + " = EXCLUDED." + mediaThumbnail + ", " +
			mediaExif + " = EXCLUDED." + mediaExif)

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
func (w *Writer) DeleteAlbum(ctx context.Context, id string) error {
	// Build the delete statement
	stmt := psql.Delete(albumsTable).
		Where(sq.Eq{albumID: id})

	// Convert to SQL
	sql, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	// Execute the query
	_, err = w.tx.Exec(ctx, sql, args...)
	return err
}

// DeleteMedia deletes a media item from the database
func (w *Writer) DeleteMedia(ctx context.Context, id string) error {
	// Build the delete statement
	stmt := psql.Delete(mediaTable).
		Where(sq.Eq{mediaID: id})

	// Convert to SQL
	sql, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	// Execute the query
	_, err = w.tx.Exec(ctx, sql, args...)
	return err
}

func (w *Writer) WriteToken(ctx context.Context, token string) error {
	stmt := tokenWriteStmt.
		Values(1, token).
		Suffix("ON CONFLICT (id) DO UPDATE SET token = excluded.token;")

	sql, args, err := stmt.ToSql()
	if err != nil {
		return err
	}

	_, err = w.tx.Exec(ctx, sql, args...)
	return err
}

func (w *Writer) AcquireGlobalLock(ctx context.Context) error {
	return w.acquireLock(ctx, false)
}

func (w *Writer) AcquireSharedLock(ctx context.Context) error {
	return w.acquireLock(ctx, true)
}

// AcquireLock attempts to acquire either a shared or global advisory lock
func (w *Writer) acquireLock(ctx context.Context, isShared bool) error {
	h := fnv.New32a()
	h.Write([]byte(lockKey))
	lockID := int32(h.Sum32())

	lockStmt := globalLockStmt
	if isShared {
		lockStmt = sharedLockStmt
	}

	_, err := w.tx.Exec(ctx, fmt.Sprintf(lockStmt, lockID))
	if err != nil {
		return fmt.Errorf("lock query failed: %w", err)
	}

	return nil
}
