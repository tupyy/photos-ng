package pg

// copyright SpiceDB

import (
	"context"

	sq "github.com/Masterminds/squirrel"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg/models"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/datastore"
)

type QueryOption func(original sq.SelectBuilder) sq.SelectBuilder

type TxUserFunc func(context.Context, *Writer) error

type Datastore struct {
	pool datastore.ConnPooler
}

// NewPostgresDatastore creates a new Postgres datastore instance with the given configuration options.
// It establishes a connection pool and sets up query interceptors for logging and monitoring.
func NewPostgresDatastore(pool datastore.ConnPooler) *Datastore {
	return &Datastore{pool: pool}
}

func (d *Datastore) QueryAlbum(ctx context.Context, opts ...QueryOption) (*entity.Album, error) {
	query := listAlbumStmt
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
		var (
			album models.Album
			err   error
		)

		err = rows.Scan(
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

	if len(albums) == 0 {
		return nil, nil
	}

	return &albums.Entity()[0], nil
}

func (d *Datastore) QueryAlbums(ctx context.Context, opts ...QueryOption) ([]entity.Album, error) {
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
		var (
			album models.Album
			err   error
		)

		err = rows.Scan(
			&album.ID,
			&album.CreatedAt,
			&album.Path,
			&album.Description,
			&album.ParentID,
			&album.ThumbnailID,
			&album.MediaCount,
			&album.ChildID,
			&album.ChildCreatedAt,
			&album.ChildPath,
			&album.ChildDescription,
			&album.ChildThumbnailID,
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
			&media.Thumbnail,
			&media.Hash,
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

func (d *Datastore) CountAlbums(ctx context.Context, opts ...QueryOption) (int, error) {
	// Start with base count query
	query := psql.Select("COUNT(*)").From(albumsTable)

	// Apply query options (filters)
	for _, opt := range opts {
		query = opt(query)
	}

	// Build and execute the query
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var count int
	err = d.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (d *Datastore) Stats(ctx context.Context) (entity.Stats, error) {
	var stats entity.Stats

	// First query: Get album and media counts
	row := d.pool.QueryRow(ctx, statAlbumMediaStmt)
	err := row.Scan(&stats.CountAlbum, &stats.CountMedia)
	if err != nil {
		return entity.Stats{}, err
	}

	// Second query: Get years with at least one image
	rows, err := d.pool.Query(ctx, statYearsStmt)
	if err != nil {
		return entity.Stats{}, err
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var year int
		if err := rows.Scan(&year); err != nil {
			return entity.Stats{}, err
		}
		years = append(years, year)
	}

	if err := rows.Err(); err != nil {
		return entity.Stats{}, err
	}

	stats.Years = years
	return stats, nil
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
