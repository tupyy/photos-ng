package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path"
	"slices"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// MediaService provides business logic for media operations
type MediaService struct {
	dt     *pg.Datastore
	fs     *fs.Datastore
	logger *logger.StructuredLogger
}

// NewMediaService creates a new instance of MediaService with the provided datastores
func NewMediaService(dt *pg.Datastore, fsDatastore *fs.Datastore) *MediaService {
	return &MediaService{
		dt:     dt,
		fs:     fsDatastore,
		logger: logger.New("media_service"),
	}
}

// GetMedia retrieves a list of media items based on the provided filter criteria
func (m *MediaService) GetMedia(ctx context.Context, filter *MediaOptions) ([]entity.Media, *PaginationCursor, error) {
	logger := m.logger.WithContext(ctx).Debug("get_media").
		WithStringPtr(AlbumID, filter.AlbumID).
		WithParam("start_date", filter.StartDate).
		WithParam("end_date", filter.EndDate).
		WithInt("filter_count", len(filter.QueriesFn())).
		Build()

	// Store original limit and request one extra for cursor generation
	originalLimit := filter.MediaLimit
	filter.MediaLimit = originalLimit + 1

	// If cursor is provided, look up the actual timestamp for accurate pagination
	if filter.Cursor != nil {
		actualMedia, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(filter.Cursor.ID), pg.Limit(1))
		if err == nil && len(actualMedia) > 0 {
			// Update cursor with actual timestamp from database
			originalTime := filter.Cursor.CapturedAt
			filter.Cursor.CapturedAt = actualMedia[0].CapturedAt
			logger.Step("cursor timestamp corrected").
				WithString("cursor_id", filter.Cursor.ID).
				WithParam("original_time", originalTime).
				WithParam("corrected_time", filter.Cursor.CapturedAt).
				Log()
		} else {
			logger.Step("cursor media not found").
				WithString("cursor_id", filter.Cursor.ID).
				Log()
		}
	}

	// Database query with debug timing
	logger.Step("database_query").
		WithString("query_type", "list_media").
		WithInt("filters", len(filter.QueriesFn())).
		Log()

	media, err := m.dt.QueryMedia(ctx, filter.QueriesFn()...)
	// Debug performance info (not error logging)
	if err != nil {
		// Return ServiceError (handlers will log the error)
		return nil, nil, NewDatabaseWriteError(ctx, "get_media", err).
			AtStep("query_media")
	}

	// Reverse results for backward direction to maintain DESC timeline order
	if filter.Direction == "backward" {
		logger.Step("reverse_results").
			WithInt("total_items", len(media)).
			WithString("direction", "backward").
			Log()

		// Reverse the slice to maintain chronological order (newest first)
		slices.Reverse(media)

		logger.Step("reversed results for backward navigation").
			WithInt("total_items", len(media)).
			Log()
	}

	// Apply date filtering in-memory for now
	// TODO: Move this to database-level filtering
	if filter.StartDate != nil || filter.EndDate != nil {
		logger.Step("date_filtering").
			WithInt("total_before_filter", len(media)).
			WithBool("has_start_date", filter.StartDate != nil).
			WithBool("has_end_date", filter.EndDate != nil).
			Log()

		filteredMedia := make([]entity.Media, 0, len(media))
		for _, item := range media {
			if filter.StartDate != nil && item.CapturedAt.Before(*filter.StartDate) {
				continue
			}
			if filter.EndDate != nil && item.CapturedAt.After(*filter.EndDate) {
				continue
			}
			filteredMedia = append(filteredMedia, item)
		}

		logger.Step("applied date filtering").
			WithInt("total_before", len(media)).
			WithInt("total_after", len(filteredMedia)).
			WithInt("filtered_out", len(media)-len(filteredMedia)).
			Log()

		media = filteredMedia
	}

	// Generate next cursor if we have more items than requested
	var nextCursor *PaginationCursor
	if len(media) > originalLimit {
		extraItem := media[originalLimit]
		nextCursor = &PaginationCursor{
			CapturedAt: extraItem.CapturedAt,
			ID:         extraItem.ID,
		}
		// Trim results to original limit for return
		media = media[:originalLimit]
	}

	logger.Success().
		WithInt(MediaReturned, len(media)).
		WithBool(DateFiltered, filter.StartDate != nil || filter.EndDate != nil).
		WithBool("has_next_cursor", nextCursor != nil).
		Log()

	return media, nextCursor, nil
}

// GetMediaByID retrieves a specific media item by its ID
func (m *MediaService) GetMediaByID(ctx context.Context, id string) (*entity.Media, error) {
	logger := m.logger.WithContext(ctx).Debug("get_media_by_id").
		WithString(MediaID, id).
		Build()

	// Input validation (return ServiceError, no logging)
	if id == "" {
		return nil, NewValidationError(ctx, "get_media", "invalid_input").
			WithContext("validation_error", "empty_media_id")
	}

	// Database query with debug timing
	logger.Step("database_query").
		WithString("query_type", "single_media").
		WithInt("filters", 2).
		Log()

	media, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(id), pg.Limit(1))
	// Debug performance info (not error logging)
	if err != nil {
		// Return ServiceError (handlers will log the error)
		return nil, NewDatabaseWriteError(ctx, "get_media", err).
			WithMediaID(id).
			AtStep("query_media")
	}

	if len(media) == 0 {
		// Return ServiceError (handlers will log the error)
		return nil, NewMediaNotFoundError(ctx, id)
	}

	processedMedia := media[0]

	// Populate the content function using the filesystem datastore
	logger.Step("content_function_setup").
		WithString(Filepath, processedMedia.Filepath()).
		Log()

	processedMedia.Content = m.fs.Read(ctx, processedMedia.Filepath())

	logger.Success().
		WithString(MediaID, processedMedia.ID).
		WithString(Filename, processedMedia.Filename).
		WithString(Filepath, processedMedia.Filepath()).
		Log()

	return &processedMedia, nil
}

// WriteMedia creates or updates a media item and writes its content to disk
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	logger := m.logger.WithContext(ctx).Debug("write_media").
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		WithString(AlbumID, media.Album.ID).
		WithString("album_path", media.Album.Path).
		Build()

	// Check if the media already exists
	logger.Step("existence_check").
		WithString("checking", "media_exists").
		Log()

	oldMedia, err := m.GetMediaByID(ctx, media.ID)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			logger.Step("media does not exist, proceeding with creation").
				WithString(MediaID, media.ID).
				Log()
		default:
			return nil, NewInternalError(ctx, "write_media", "check_existing", err).
				WithMediaID(media.ID)
		}
	}

	// Content reading with logging
	logger.Step("content_reading").
		WithString(Filename, media.Filename).
		Log()

	content, err := media.Content()
	if err != nil {
		return nil, NewMediaProcessingError(ctx, "read_content", media.Filename, err)
	}

	// Read all content into memory to compute hash and write to disk
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, NewMediaProcessingError(ctx, "read_content_bytes", media.Filename, err)
	}

	// Compute SHA256 hash
	logger.Step("hash_computation").
		WithInt("content_size", len(contentBytes)).
		Log()

	hash := sha256.Sum256(contentBytes)
	hashStr := fmt.Sprintf("%x", hash)

	if oldMedia != nil && hashStr == oldMedia.Hash {
		logger.Success().
			WithString(MediaID, oldMedia.ID).
			WithString(Filename, oldMedia.Filename).
			WithString(Hash, hashStr).
			WithBool("skipped", true).
			WithString("reason", "unchanged_content").
			Log()
		return oldMedia, nil
	}

	// Transaction with detailed step logging
	logger.Step("transaction_start").
		WithInt("content_size", len(contentBytes)).
		WithString(Hash, hashStr).
		WithBool("is_update", oldMedia != nil).
		Log()

	logger.Step("starting").
		WithString("operation", "write_media").
		Log()

	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		media.Hash = hashStr

		// Step 1: Initialize processing service
		logger.Step("processing_init").
			WithString(Filename, media.Filename).
			Log()

		processingSrv, err := NewProcessingMediaService()
		if err != nil {
			return NewMediaProcessingError(ctx, "init_processing", media.Filename, err)
		}

		// Step 2: Process media (thumbnail + EXIF)
		logger.Step("media_processing").
			WithString(Filename, media.Filename).
			WithInt("content_size", len(contentBytes)).
			Log()

		procStart := time.Now()
		r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
		processingDuration := time.Since(procStart)
		if err != nil {
			return NewMediaProcessingError(ctx, "generate_thumbnail", media.Filename, err)
		}

		logger.Step("media_processing").
			WithString("filename", media.Filename).
			WithInt("content_size", len(contentBytes)).
			WithInt("exif_fields", len(exif)).
			WithParam("duration", processingDuration).
			Log()

		// Step 3: Read thumbnail
		thumbnail, err := io.ReadAll(r)
		if err != nil {
			return NewMediaProcessingError(ctx, "read_thumbnail", media.Filename, err)
		}
		media.Thumbnail = thumbnail
		media.Exif = exif

		logger.Step("thumbnail_generated").
			WithString("filename", media.Filename).
			WithInt("thumbnail_size", len(thumbnail)).
			WithInt("exif_fields", len(exif)).
			Log()

		// Step 4: Extract capture time
		if captureAt, err := media.GetCapturedTime(); err != nil {
			logger.Step("capture_time_extraction_failed").
				WithString("filename", media.Filename).
				WithString("error", err.Error()).
				Log()
		} else {
			media.CapturedAt = captureAt
			logger.Step("capture_time_extracted").
				WithString("filename", media.Filename).
				WithParam(CapturedAt, captureAt).
				Log()
		}

		// Step 5: Write file to disk
		logger.Step("filesystem_write").
			WithString(Filepath, media.Filepath()).
			Log()

		if err := m.fs.Write(ctx, media.Filepath(), bytes.NewReader(contentBytes)); err != nil {
			return NewFilesystemError(ctx, "write_media", "filesystem_write", media.Filepath(), err)
		}

		// Step 6: Write to database
		logger.Step("database_write").
			WithString("table", "media").
			Log()

		if err := writer.WriteMedia(ctx, media); err != nil {
			return NewDatabaseWriteError(ctx, "write_media", err).
				WithMediaID(media.ID).
				WithFilename(media.Filename)
		}

		return nil
	})
	if err != nil {
		return nil, NewInternalError(ctx, "write_media", "transaction", err).
			WithMediaID(media.ID).
			WithFilename(media.Filename)
	}

	logger.Step("completed").
		WithString("operation", "write_media").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		WithString(Hash, hashStr).
		WithInt("content_size", len(contentBytes)).
		WithInt("thumbnail_size", len(media.Thumbnail)).
		WithInt("exif_fields", len(media.Exif)).
		WithParam(CapturedAt, media.CapturedAt).
		Log()

	return &media, nil
}

// UpdateMedia updates an existing media item using an entity.Media
func (m *MediaService) UpdateMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	logger := m.logger.WithContext(ctx).Debug("update_media").
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		Build()

	// Clear the content function to avoid writing file content during update
	media.Content = nil

	logger.Step("metadata-only update, content function cleared").
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		Log()

	// Database transaction with debug timing
	logger.Step("database_update").
		WithString("table", "media").
		Log()

	logger.Step("starting").
		WithString("operation", "update_media").
		Log()

	err := m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteMedia(ctx, media)
	})
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "update_media", err).
			WithMediaID(media.ID).
			WithFilename(media.Filename)
	}

	logger.Step("completed").
		WithString("operation", "update_media").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		Log()

	return &media, nil
}

// DeleteMedia deletes a media item by ID
func (m *MediaService) DeleteMedia(ctx context.Context, id string) error {
	logger := m.logger.WithContext(ctx).Debug("delete_media").
		WithString(MediaID, id).
		Build()

	// Check if media exists
	logger.Step("validate_exists").
		WithString(MediaID, id).
		Log()

	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return NewInternalError(ctx, "delete_media", "validate_exists", err).
			WithMediaID(id)
	}

	logger.Step("media found, proceeding with deletion").
		WithString(MediaID, media.ID).
		WithString(Filename, media.Filename).
		WithString(Filepath, media.Filepath()).
		Log()

	// Delete the media from the datastore using a write transaction
	logger.Step("transaction_start").
		WithString(Filepath, media.Filepath()).
		WithParam("operations", []string{"filesystem_delete", "database_delete"}).
		Log()

	logger.Step("starting").
		WithString("operation", "delete_media").
		Log()

	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Remove the file from album folders
		logger.Step("filesystem_delete").
			WithString(Filepath, media.Filepath()).
			Log()

		if err := m.fs.DeleteMedia(ctx, media.Filepath()); err != nil {
			return NewFilesystemError(ctx, "delete_media", "filesystem_delete", media.Filepath(), err)
		}

		// Delete the media from the database
		logger.Step("database_delete").
			WithString("table", "media").
			WithString(MediaID, id).
			Log()

		if err := writer.DeleteMedia(ctx, id); err != nil {
			return NewDatabaseWriteError(ctx, "delete_media", err).
				WithMediaID(id)
		}

		return nil
	})
	if err != nil {
		return NewInternalError(ctx, "delete_media", "transaction", err).
			WithMediaID(id).
			WithFilepath(media.Filepath())
	}

	logger.Step("completed").
		WithString("operation", "delete_media").
		WithBool("success", true).
		Log()

	logger.Success().
		WithString(MediaID, id).
		WithString(Filepath, media.Filepath()).
		WithBool("filesystem_deleted", true).
		WithBool("database_deleted", true).
		Log()

	return nil
}

func (m *MediaService) GetContentFn(ctx context.Context, media entity.Media) entity.MediaContentFn {
	// Construct the full file path from media filename and album path
	filepath := path.Join(media.Album.Path, media.Filename)
	return m.fs.Read(ctx, filepath)
}
