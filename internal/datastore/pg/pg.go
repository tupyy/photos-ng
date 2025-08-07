package pg

// copyright SpiceDB

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg/models"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryOption func(original sq.SelectBuilder) sq.SelectBuilder

type TxUserFunc func(context.Context, *Writer) error

type Datastore struct {
	pool ConnPooler
}

// NewPostgresDatastore creates a new Postgres datastore instance with the given configuration options.
// It establishes a connection pool and sets up query interceptors for logging and monitoring.
func NewPostgresDatastore(ctx context.Context, url string, options ...Option) (*Datastore, error) {
	pgOptions := newPostgresConfig(options)

	pgxConfig, err := pgOptions.PgxConfig(url)
	if err != nil {
		return nil, err
	}

	pgPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	if err := pgPool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Datastore{pool: MustNewInterceptorPooler(pgPool, newLogInterceptor())}, nil
}

func (d *Datastore) QueryAlbums(ctx context.Context, opts ...QueryOption) ([]entity.Album, error) {
	// Start with the base listAlbumsStmt and apply any query options
	query := listAlbumsStmt
	for _, opt := range opts {
		query = opt(query)
	}

	// Build the SQL query
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query
	rows, err := d.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan results into album models
	albums := models.Albums{}
	for rows.Next() {
		var album models.Album
		err := rows.Scan(
			&album.ID,
			&album.CreatedAt,
			&album.Path,
			&album.Description,
			&album.ParentID,
			&album.ThumbnailID,
			&album.ChildID,
			&album.ChildCreatedAt,
			&album.ChildPath,
			&album.ChildDescription,
			&album.ChildThumbnailID,
			&album.MediaID,
			&album.MediaCapturedAt,
			&album.MediaAlbumID,
			&album.MediaFileName,
			&album.MediaThumbnail,
			&album.MediaExif,
			&album.MediaMediaType,
		)
		if err != nil {
			return nil, err
		}
		albums.Add(album)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert to entity albums
	return albums.Entity(), nil
}

func (d *Datastore) QueryMedia(ctx context.Context, opts ...QueryOption) ([]entity.Media, error) {
	// Start with the base listMediaStmt and apply any query options
	query := listMediaStmt
	for _, opt := range opts {
		query = opt(query)
	}

	// Build the SQL query
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query
	rows, err := d.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan results into media models
	mediaList := models.MediaList{}
	for rows.Next() {
		var media models.Media
		err := rows.Scan(
			&media.ID,
			&media.CreatedAt,
			&media.CapturedAt,
			&media.AlbumID,
			&media.FileName,
			&media.Hash,
			&media.Thumbnail,
			&media.Exif,
			&media.MediaType,
			&media.AlbumJoinCreatedAt,
			&media.AlbumJoinPath,
			&media.AlbumJoinDescription,
			&media.AlbumJoinThumbnailID,
		)
		if err != nil {
			return nil, err
		}
		mediaList.Add(media)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert to entity media
	return mediaList.Entity(), nil
}

func (d *Datastore) Stats(ctx context.Context) (entity.Stats, error) {
	// TODO: Implement actual stats query
	return entity.Stats{
		CountMedia:    0,
		CountAlbum:    0,
		TimelineYears: []int{}, // This should be populated from actual media data
	}, nil
}

// WriteTx executes a write transaction with the provided user function.
// It manages transaction lifecycle and provides a Writer interface for data modifications.
func (d *Datastore) WriteTx(ctx context.Context, txFn TxUserFunc) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}

	writer := &Writer{tx: tx}

	if err := txFn(ctx, writer); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (dt *Datastore) Close() {
	dt.pool.Close()
}
