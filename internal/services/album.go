package services

import (
	"context"
	"path"
	"strings"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// AlbumService provides business logic for album operations
type AlbumService struct {
	dt     *pg.Datastore
	fs     *fs.Datastore
	logger *logger.StructuredLogger
}

// NewAlbumService creates a new instance of AlbumService with the provided datastores
func NewAlbumService(dt *pg.Datastore, fsDatastore *fs.Datastore) *AlbumService {
	return &AlbumService{
		dt:     dt,
		fs:     fsDatastore,
		logger: logger.New("album_service"),
	}
}

func (a *AlbumService) List(ctx context.Context, opts *ListOptions) ([]entity.Album, error) {
	logger := a.logger.WithContext(ctx).Debug("get_albums").
		WithInt("limit", opts.Limit).
		WithInt("offset", opts.Offset).
		WithBool("has_parent", opts.HasParent).
		WithInt("filter_count", len(opts.QueriesFn())).
		Build()

	logger.Step("database_query").
		WithString("query_type", "list_albums").
		WithInt("filters", len(opts.QueriesFn())).
		Log()

	allAlbums, err := a.dt.QueryAlbums(ctx, opts.QueriesFn()...)
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "get_albums", err).
			AtStep("query_albums")
	}

	// Apply pagination at the application level
	logger.Step("pagination").
		WithInt("total_albums", len(allAlbums)).
		WithInt("start", opts.Offset).
		WithInt("end", opts.Offset+opts.Limit).
		Log()

	startIdx := opts.Offset
	endIdx := opts.Offset + opts.Limit

	// Handle bounds
	if startIdx >= len(allAlbums) {
		logger.Step("pagination out of bounds, returning empty result").
			WithInt("start_index", startIdx).
			WithInt("total_albums", len(allAlbums)).
			Log()
		logger.Success().
			WithInt(AlbumsReturned, 0).
			WithInt(TotalAlbums, len(allAlbums)).
			Log()
		return []entity.Album{}, nil
	}
	if endIdx > len(allAlbums) || opts.Limit <= 0 {
		endIdx = len(allAlbums)
		logger.Step("pagination end adjusted to total albums").
			WithInt("original_end", opts.Offset+opts.Limit).
			WithInt("adjusted_end", endIdx).
			WithInt("total_albums", len(allAlbums)).
			Log()
	}

	// Return the paginated slice
	paginatedAlbums := allAlbums[startIdx:endIdx]

	logger.Success().
		WithInt(AlbumsReturned, len(paginatedAlbums)).
		WithInt(TotalAlbums, len(allAlbums)).
		WithInt(StartIndex, startIdx).
		WithInt(EndIndex, endIdx).
		Log()

	return paginatedAlbums, nil
}

func (a *AlbumService) Count(ctx context.Context, hasParent bool) (int, error) {
	logger := a.logger.WithContext(ctx).Debug("count_albums").
		WithBool("has_parent", hasParent).
		Build()

	logger.Step("database_count").
		WithString("query_type", "count_albums").
		Log()

	count, err := a.dt.CountAlbums(ctx, pg.FilterAlbumWithParents(hasParent))
	if err != nil {
		return 0, NewDatabaseWriteError(ctx, "count_albums", err).
			AtStep("count_albums")
	}

	logger.Success().
		WithInt("total_albums", count).
		Log()

	return count, nil
}

func (a *AlbumService) Get(ctx context.Context, id string) (*entity.Album, error) {
	logger := a.logger.WithContext(ctx).Debug("get_album").
		WithString(AlbumID, id).
		Build()

	if id == "" {
		return nil, NewValidationError(ctx, "get_album", "invalid_input").
			WithContext("validation_error", "empty_album_id")
	}

	logger.Step("database_query").
		WithString("query_type", "single_album").
		WithInt("filters", 1).
		Log()

	album, err := a.dt.QueryAlbum(ctx, pg.FilterByAlbumId(id))
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "get_album", err).
			WithAlbumID(id).
			AtStep("query_album")
	}

	if album == nil {
		return nil, NewAlbumNotFoundError(ctx, id)
	}

	logger.Success().
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		Log()

	return album, nil
}

// CreateAlbum creates a new album using an entity.Album
func (a *AlbumService) Create(ctx context.Context, album entity.Album) (*entity.Album, error) {
	logger := a.logger.WithContext(ctx).Debug("create_album").
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		WithStringPtr("parent_id", album.ParentId).
		Build()

	logger.Step("existence_check").
		WithString("checking", "album_exists").
		Log()

	isAlbumExists := true
	if _, err := a.Get(ctx, album.ID); err != nil {
		switch err.(type) {
		case *NotFoundError:
			isAlbumExists = false
			logger.Step("album does not exist, proceeding with creation").
				WithString(AlbumID, album.ID).
				Log()
		default:
			return nil, NewInternalError(ctx, "create_album", "check_album_exists", err).
				WithAlbumID(album.ID)
		}
	} else {
		return nil, NewAlbumExistsError(ctx, album.ID, album.Path)
	}

	if album.ParentId != nil {
		logger.Step("parent_processing").
			WithString("parent_id", *album.ParentId).
			Log()

		parent, err := a.Get(ctx, *album.ParentId)
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
			logger.Step("album path computed for parent relationship").
				WithString("original_path", originalPath).
				WithString("computed_path", album.Path).
				WithString("parent_path", parent.Path).
				Log()
		}
		album.ID = entity.GenerateId(album.Path)
	}

	// Create the album in the datastore using a write transaction
	logger.Step("transaction_start").
		WithString("final_album_id", album.ID).
		WithString("final_album_path", album.Path).
		WithBool("is_new_album", !isAlbumExists).
		Log()

	logger.Step("starting").
		WithString("operation", "create_album").
		Log()

	err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Write the album to database
		logger.Step("database_write").
			WithString("table", "albums").
			Log()

		if err := writer.WriteAlbum(ctx, album); err != nil {
			return NewDatabaseWriteError(ctx, "create_album", err).
				WithAlbumID(album.ID).
				WithAlbumPath(album.Path)
		}

		// If it's a new album, create the folder on disk
		if !isAlbumExists {
			logger.Step("filesystem_create").
				WithString("folder_path", album.Path).
				Log()

			if err := a.fs.CreateFolder(ctx, album.Path); err != nil {
				return NewFilesystemError(ctx, "create_album", "filesystem_create", album.Path, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, NewInternalError(ctx, "create_album", "transaction", err).
			WithAlbumID(album.ID).
			WithAlbumPath(album.Path)
	}

	logger.Step("completed").
		WithString("operation", "create_album").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		WithBool(WasExisting, isAlbumExists).
		Log()

	return &album, nil
}

// UpdateAlbum updates an existing album
func (a *AlbumService) Update(ctx context.Context, album entity.Album) (*entity.Album, error) {
	logger := a.logger.WithContext(ctx).Debug("update_album").
		WithString(AlbumID, album.ID).
		WithStringPtr("description", album.Description).
		WithStringPtr("thumbnail", album.Thumbnail).
		Build()

	// Validate album exists
	logger.Step("validate_exists").
		WithString(AlbumID, album.ID).
		Log()

	existingAlbum, err := a.Get(ctx, album.ID)
	if err != nil {
		return nil, NewInternalError(ctx, "update_album", "validate_exists", err).
			WithAlbumID(album.ID)
	}

	logger.Step("applying updates to existing album").
		WithStringPtr("existing_description", existingAlbum.Description).
		WithStringPtr("new_description", album.Description).
		WithBool("has_thumbnail_update", album.Thumbnail != nil).
		Log()

	existingAlbum.Description = album.Description

	// if thumbnail is present, check if the media belongs to the album
	if album.Thumbnail != nil {
		logger.Step("validate_thumbnail").
			WithString("thumbnail_id", *album.Thumbnail).
			Log()

		media, err := a.dt.QueryMedia(ctx, pg.FilterByMediaId(*album.Thumbnail), pg.Limit(1))
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

		logger.Step("thumbnail validation successful").
			WithString("thumbnail_id", *album.Thumbnail).
			WithBool("media_found", true).
			Log()

		existingAlbum.Thumbnail = album.Thumbnail
	}

	// Write the album in the datastore using a write transaction
	logger.Step("database_update").
		WithString("table", "albums").
		Log()

	logger.Step("starting").
		WithString("operation", "update_album").
		Log()

	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteAlbum(ctx, *existingAlbum)
	})
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "update_album", err).
			WithAlbumID(album.ID)
	}

	logger.Step("completed").
		WithString("operation", "update_album").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(AlbumID, album.ID).
		WithBool(DescriptionUpdated, true).
		WithBool(ThumbnailUpdated, album.Thumbnail != nil).
		Log()

	return existingAlbum, nil
}

func (a *AlbumService) Delete(ctx context.Context, id string) error {
	logger := a.logger.WithContext(ctx).Debug("delete_album").
		WithString(AlbumID, id).
		Build()

	// Check if album exists
	logger.Step("validate_exists").
		WithString(AlbumID, id).
		Log()

	album, err := a.Get(ctx, id)
	if err != nil {
		return NewInternalError(ctx, "delete_album", "validate_exists", err).
			WithAlbumID(id)
	}

	logger.Step("album found, proceeding with deletion").
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		Log()

	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Delete the album folder from the file system
		logger.Step("filesystem_delete").
			WithString("folder_path", album.Path).
			Log()

		if err := a.fs.DeleteFolder(ctx, album.Path); err != nil {
			return NewFilesystemError(ctx, "delete_album", "filesystem_delete", album.Path, err)
		}

		// Delete the album from the database
		logger.Step("database_delete").
			WithString("table", "albums").
			WithString(AlbumID, id).
			Log()

		if err := writer.DeleteAlbum(ctx, id); err != nil {
			return NewDatabaseWriteError(ctx, "delete_album", err).
				WithAlbumID(id)
		}

		return nil
	})
	if err != nil {
		return NewInternalError(ctx, "delete_album", "transaction", err).
			WithAlbumID(id).
			WithAlbumPath(album.Path)
	}

	logger.Step("completed").
		WithString("operation", "delete_album").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(AlbumID, id).
		WithString(AlbumPath, album.Path).
		WithBool(FilesystemDeleted, true).
		WithBool(DatabaseDeleted, true).
		Log()

	return nil
}
