package services

import (
	"context"
	"path"
	"strings"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// AlbumService provides business logic for album operations
type AlbumService struct {
	dt    *pg.Datastore
	fs    *fs.Datastore
	debug *logger.DebugLogger
}

// NewAlbumService creates a new instance of AlbumService with the provided datastores
func NewAlbumService(dt *pg.Datastore, fsDatastore *fs.Datastore) *AlbumService {
	return &AlbumService{
		dt:    dt,
		fs:    fsDatastore,
		debug: logger.NewDebugLogger("album_service"),
	}
}

// GetAlbums retrieves a list of albums based on the provided filter criteria
// Handles pagination at the application level to avoid issues with JOIN queries
func (a *AlbumService) GetAlbums(ctx context.Context, opts *AlbumOptions) ([]entity.Album, error) {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("get_albums").
		WithInt("limit", opts.Limit).
		WithInt("offset", opts.Offset).
		WithBool("has_parent", opts.HasParent).
		WithInt("filter_count", len(opts.QueriesFn())).
		Build()

	// Database query with debug timing
	tracer.Step("database_query").
		WithString("query_type", "list_albums").
		WithInt("filters", len(opts.QueriesFn())).
		Log()

	start := time.Now()
	allAlbums, err := a.dt.QueryAlbums(ctx, opts.QueriesFn()...)
	queryDuration := time.Since(start)

	// Debug performance info (not error logging)
	debug.DatabaseQuery("query_albums", len(opts.QueriesFn()), queryDuration, len(allAlbums) > 0)

	if err != nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewDatabaseWriteError(ctx, "get_albums", err).
			AtStep("query_albums")
	}

	// Apply pagination at the application level
	tracer.Step("pagination").
		WithInt("total_albums", len(allAlbums)).
		WithInt("start", opts.Offset).
		WithInt("end", opts.Offset+opts.Limit).
		Log()

	startIdx := opts.Offset
	endIdx := opts.Offset + opts.Limit

	// Handle bounds
	if startIdx >= len(allAlbums) {
		debug.BusinessLogic("pagination out of bounds, returning empty result").
			WithInt("start_index", startIdx).
			WithInt("total_albums", len(allAlbums)).
			Log()
		tracer.Success().
			WithInt(AlbumsReturned, 0).
			WithInt(TotalAlbums, len(allAlbums)).
			Log()
		return []entity.Album{}, nil
	}
	if endIdx > len(allAlbums) || opts.Limit <= 0 {
		endIdx = len(allAlbums)
		debug.BusinessLogic("pagination end adjusted to total albums").
			WithInt("original_end", opts.Offset+opts.Limit).
			WithInt("adjusted_end", endIdx).
			WithInt("total_albums", len(allAlbums)).
			Log()
	}

	// Return the paginated slice
	paginatedAlbums := allAlbums[startIdx:endIdx]

	tracer.Success().
		WithInt(AlbumsReturned, len(paginatedAlbums)).
		WithInt(TotalAlbums, len(allAlbums)).
		WithInt(StartIndex, startIdx).
		WithInt(EndIndex, endIdx).
		Log()

	return paginatedAlbums, nil
}

// CountAlbums returns the total count of albums matching the provided filter criteria
func (a *AlbumService) CountAlbums(ctx context.Context, opts *AlbumOptions) (int, error) {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("count_albums").
		WithBool("has_parent", opts.HasParent).
		WithInt("filter_count", len(opts.QueriesFn())).
		Build()

	// Database count query with debug timing
	tracer.Step("database_count").
		WithString("query_type", "count_albums").
		WithInt("filters", len(opts.QueriesFn())).
		Log()

	count, err := a.dt.CountAlbums(ctx, opts.QueriesFn()...)
	if err != nil {
		return 0, NewDatabaseWriteError(ctx, "count_albums", err).
			AtStep("count_albums")
	}

	tracer.Success().
		WithInt("total_albums", count).
		Log()

	return count, nil
}

// GetAlbum retrieves a specific album by its ID
func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("get_album").
		WithString("album_id", id).
		Build()

	// Input validation (return ServiceError, no logging)
	if id == "" {
		return nil, NewValidationError(ctx, "get_album", "invalid_input").
			WithContext("validation_error", "empty_album_id")
	}

	// Database query with debug timing
	tracer.Step("database_query").
		WithString("query_type", "single_album").
		WithInt("filters", 1).
		Log()

	start := time.Now()
	album, err := a.dt.QueryAlbum(ctx, pg.FilterByAlbumId(id))
	queryDuration := time.Since(start)

	// Debug performance info (not error logging)
	debug.DatabaseQuery("query_album", 1, queryDuration, album != nil)

	if err != nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewDatabaseWriteError(ctx, "get_album", err).
			WithAlbumID(id).
			AtStep("query_album")
	}

	if album == nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewAlbumNotFoundError(ctx, id)
	}

	// Debug success info (duration computed automatically)
	tracer.Success().
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		Log()

	return album, nil
}

// CreateAlbum creates a new album using an entity.Album
func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("create_album").
		WithString("album_id", album.ID).
		WithString("album_path", album.Path).
		WithStringPtr("parent_id", album.ParentId).
		Build()

	// Check if the album already exists
	tracer.Step("existence_check").
		WithString("checking", "album_exists").
		Log()

	isAlbumExists := true
	if _, err := a.GetAlbum(ctx, album.ID); err != nil {
		switch err.(type) {
		case *NotFoundError:
			isAlbumExists = false
			debug.BusinessLogic("album does not exist, proceeding with creation").
				WithString("album_id", album.ID).
				Log()
		default:
			return nil, NewInternalError(ctx, "create_album", "check_album_exists", err).
				WithAlbumID(album.ID)
		}
	} else {
		return nil, NewAlbumExistsError(ctx, album.ID, album.Path)
	}

	// Get parent if parentID exists and recompute path and id of the new album.
	if album.ParentId != nil {
		tracer.Step("parent_processing").
			WithString("parent_id", *album.ParentId).
			Log()

		parent, err := a.GetAlbum(ctx, *album.ParentId)
		if err != nil {
			switch err.(type) {
			case *NotFoundError:
				return nil, NewParentAlbumNotFoundError(ctx, *album.ParentId)
			default:
				return nil, NewInternalError(ctx, "create_album", "validate_parent", err).
					WithParentID(*album.ParentId)
			}
		}

		// Path computation business logic
		originalPath := album.Path
		if !strings.HasPrefix(album.Path, parent.Path+"/") && album.Path != parent.Path {
			album.Path = path.Join(parent.Path, album.Path)
			debug.BusinessLogic("album path computed for parent relationship").
				WithString("original_path", originalPath).
				WithString("computed_path", album.Path).
				WithString("parent_path", parent.Path).
				Log()
		}
		album.ID = entity.GenerateId(album.Path)
	}

	// Create the album in the datastore using a write transaction
	tracer.Step("transaction_start").
		WithString("final_album_id", album.ID).
		WithString("final_album_path", album.Path).
		WithBool("is_new_album", !isAlbumExists).
		Log()

	debug.Transaction("starting").
		WithString("operation", "create_album").
		Log()

	err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Write the album to database
		tracer.Step("database_write").
			WithString("table", "albums").
			Log()

		start := time.Now()
		if err := writer.WriteAlbum(ctx, album); err != nil {
			return NewDatabaseWriteError(ctx, "create_album", err).
				WithAlbumID(album.ID).
				WithAlbumPath(album.Path)
		}
		tracer.Performance("database_write", time.Since(start))

		// If it's a new album, create the folder on disk
		if !isAlbumExists {
			tracer.Step("filesystem_create").
				WithString("folder_path", album.Path).
				Log()

			start = time.Now()
			if err := a.fs.CreateFolder(ctx, album.Path); err != nil {
				return NewFilesystemError(ctx, "create_album", "filesystem_create", album.Path, err)
			}
			debug.FileOperation("create_folder", album.Path, 0, time.Since(start))
		}

		return nil
	})
	if err != nil {
		return nil, NewInternalError(ctx, "create_album", "transaction", err).
			WithAlbumID(album.ID).
			WithAlbumPath(album.Path)
	}

	debug.Transaction("completed").
		WithString("operation", "create_album").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		WithBool(WasExisting, isAlbumExists).
		Log()

	return &album, nil
}

// UpdateAlbum updates an existing album
func (a *AlbumService) UpdateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("update_album").
		WithString("album_id", album.ID).
		WithStringPtr("description", album.Description).
		WithStringPtr("thumbnail", album.Thumbnail).
		Build()

	// Validate album exists
	tracer.Step("validate_exists").
		WithString("album_id", album.ID).
		Log()

	existingAlbum, err := a.GetAlbum(ctx, album.ID)
	if err != nil {
		return nil, NewInternalError(ctx, "update_album", "validate_exists", err).
			WithAlbumID(album.ID)
	}

	debug.BusinessLogic("applying updates to existing album").
		WithStringPtr("existing_description", existingAlbum.Description).
		WithStringPtr("new_description", album.Description).
		WithBool("has_thumbnail_update", album.Thumbnail != nil).
		Log()

	existingAlbum.Description = album.Description

	// if thumbnail is present, check if the media belongs to the album
	if album.Thumbnail != nil {
		tracer.Step("validate_thumbnail").
			WithString("thumbnail_id", *album.Thumbnail).
			Log()

		start := time.Now()
		media, err := a.dt.QueryMedia(ctx, pg.FilterByMediaId(*album.Thumbnail), pg.Limit(1))
		queryDuration := time.Since(start)

		debug.DatabaseQuery("query_media_for_thumbnail", 1, queryDuration, len(media) > 0)

		if err != nil {
			return nil, NewDatabaseWriteError(ctx, "update_album", err).
				WithAlbumID(album.ID).
				WithContext("thumbnail_id", *album.Thumbnail).
				AtStep("validate_thumbnail")
		}

		if len(media) == 0 {
			return nil, NewValidationError(ctx, "update_album", "thumbnail_not_found").
				WithAlbumID(album.ID).
				WithContext("thumbnail_id", *album.Thumbnail)
		}

		debug.BusinessLogic("thumbnail validation successful").
			WithString("thumbnail_id", *album.Thumbnail).
			WithBool("media_found", true).
			Log()

		existingAlbum.Thumbnail = album.Thumbnail
	}

	// Write the album in the datastore using a write transaction
	tracer.Step("database_update").
		WithString("table", "albums").
		Log()

	debug.Transaction("starting").
		WithString("operation", "update_album").
		Log()

	start := time.Now()
	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteAlbum(ctx, *existingAlbum)
	})
	transactionDuration := time.Since(start)

	tracer.Performance("transaction_duration", transactionDuration)

	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "update_album", err).
			WithAlbumID(album.ID)
	}

	debug.Transaction("completed").
		WithString("operation", "update_album").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString(AlbumID, album.ID).
		WithBool(DescriptionUpdated, true).
		WithBool(ThumbnailUpdated, album.Thumbnail != nil).
		Log()

	return existingAlbum, nil
}

// DeleteAlbum deletes an album by ID
func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
	debug := a.debug.WithContext(ctx)
	tracer := debug.StartOperation("delete_album").
		WithString("album_id", id).
		Build()

	// Check if album exists
	tracer.Step("validate_exists").
		WithString("album_id", id).
		Log()

	album, err := a.GetAlbum(ctx, id)
	if err != nil {
		return NewInternalError(ctx, "delete_album", "validate_exists", err).
			WithAlbumID(id)
	}

	debug.BusinessLogic("album found, proceeding with deletion").
		WithString("album_id", album.ID).
		WithString("album_path", album.Path).
		Log()

	// Delete the album from the datastore using a write transaction
	tracer.Step("transaction_start").
		WithString("album_path", album.Path).
		WithParam("operations", []string{"filesystem_delete", "database_delete"}).
		Log()

	debug.Transaction("starting").
		WithString("operation", "delete_album").
		Log()

	start := time.Now()
	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Delete the album folder from the file system
		tracer.Step("filesystem_delete").
			WithString("folder_path", album.Path).
			Log()

		fsStart := time.Now()
		if err := a.fs.DeleteFolder(ctx, album.Path); err != nil {
			return NewFilesystemError(ctx, "delete_album", "filesystem_delete", album.Path, err)
		}
		debug.FileOperation("delete_folder", album.Path, 0, time.Since(fsStart))

		// Delete the album from the database
		tracer.Step("database_delete").
			WithString("table", "albums").
			WithString("album_id", id).
			Log()

		dbStart := time.Now()
		if err := writer.DeleteAlbum(ctx, id); err != nil {
			return NewDatabaseWriteError(ctx, "delete_album", err).
				WithAlbumID(id)
		}
		tracer.Performance("database_delete", time.Since(dbStart))

		return nil
	})
	transactionDuration := time.Since(start)

	tracer.Performance("transaction_duration", transactionDuration)

	if err != nil {
		return NewInternalError(ctx, "delete_album", "transaction", err).
			WithAlbumID(id).
			WithAlbumPath(album.Path)
	}

	debug.Transaction("completed").
		WithString("operation", "delete_album").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString(AlbumID, id).
		WithString(AlbumPath, album.Path).
		WithBool(FilesystemDeleted, true).
		WithBool(DatabaseDeleted, true).
		Log()

	return nil
}
