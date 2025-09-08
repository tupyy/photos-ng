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
	dt    *pg.Datastore
	fs    *fs.Datastore
	debug *logger.DebugLogger
}

// NewMediaService creates a new instance of MediaService with the provided datastores
func NewMediaService(dt *pg.Datastore, fsDatastore *fs.Datastore) *MediaService {
	return &MediaService{
		dt:    dt,
		fs:    fsDatastore,
		debug: logger.NewDebugLogger("media_service"),
	}
}

// GetMedia retrieves a list of media items based on the provided filter criteria
func (m *MediaService) GetMedia(ctx context.Context, filter *MediaOptions) ([]entity.Media, *PaginationCursor, error) {
	debug := m.debug.WithContext(ctx)
	tracer := debug.StartOperation("get_media").
		WithStringPtr("album_id", filter.AlbumID).
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
			debug.BusinessLogic("cursor timestamp corrected").
				WithString("cursor_id", filter.Cursor.ID).
				WithParam("original_time", originalTime).
				WithParam("corrected_time", filter.Cursor.CapturedAt).
				Log()
		} else {
			debug.BusinessLogic("cursor media not found").
				WithString("cursor_id", filter.Cursor.ID).
				Log()
		}
	}

	// Database query with debug timing
	tracer.Step("database_query").
		WithString("query_type", "list_media").
		WithInt("filters", len(filter.QueriesFn())).
		Log()

	start := time.Now()
	media, err := m.dt.QueryMedia(ctx, filter.QueriesFn()...)
	queryDuration := time.Since(start)

	// Debug performance info (not error logging)
	debug.DatabaseQuery("query_media", len(filter.QueriesFn()), queryDuration, len(media) > 0)

	if err != nil {
		// Return ServiceError (handlers will log the error)
		return nil, nil, NewDatabaseWriteError(ctx, "get_media", err).
			AtStep("query_media")
	}

	// Reverse results for backward direction to maintain DESC timeline order
	if filter.Direction == "backward" {
		tracer.Step("reverse_results").
			WithInt("total_items", len(media)).
			WithString("direction", "backward").
			Log()

		// Reverse the slice to maintain chronological order (newest first)
		slices.Reverse(media)

		debug.BusinessLogic("reversed results for backward navigation").
			WithInt("total_items", len(media)).
			Log()
	}

	// Apply date filtering in-memory for now
	// TODO: Move this to database-level filtering
	if filter.StartDate != nil || filter.EndDate != nil {
		tracer.Step("date_filtering").
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

		debug.BusinessLogic("applied date filtering").
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

	tracer.Success().
		WithInt(MediaReturned, len(media)).
		WithBool(DateFiltered, filter.StartDate != nil || filter.EndDate != nil).
		WithBool("has_next_cursor", nextCursor != nil).
		Log()

	return media, nextCursor, nil
}

// GetMediaByID retrieves a specific media item by its ID
func (m *MediaService) GetMediaByID(ctx context.Context, id string) (*entity.Media, error) {
	debug := m.debug.WithContext(ctx)
	tracer := debug.StartOperation("get_media_by_id").
		WithString("media_id", id).
		Build()

	// Input validation (return ServiceError, no logging)
	if id == "" {
		return nil, NewValidationError(ctx, "get_media", "invalid_input").
			WithContext("validation_error", "empty_media_id")
	}

	// Database query with debug timing
	tracer.Step("database_query").
		WithString("query_type", "single_media").
		WithInt("filters", 2).
		Log()

	start := time.Now()
	media, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(id), pg.Limit(1))
	queryDuration := time.Since(start)

	// Debug performance info (not error logging)
	debug.DatabaseQuery("query_media", 2, queryDuration, len(media) > 0)

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
	tracer.Step("content_function_setup").
		WithString("filepath", processedMedia.Filepath()).
		Log()

	processedMedia.Content = m.fs.Read(ctx, processedMedia.Filepath())

	tracer.Success().
		WithString(MediaID, processedMedia.ID).
		WithString(Filename, processedMedia.Filename).
		WithString(Filepath, processedMedia.Filepath()).
		Log()

	return &processedMedia, nil
}

// WriteMedia creates or updates a media item and writes its content to disk
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	debug := m.debug.WithContext(ctx)
	tracer := debug.StartOperation("write_media").
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		WithString("album_id", media.Album.ID).
		WithString("album_path", media.Album.Path).
		Build()

	// Check if the media already exists
	tracer.Step("existence_check").
		WithString("checking", "media_exists").
		Log()

	oldMedia, err := m.GetMediaByID(ctx, media.ID)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			debug.BusinessLogic("media does not exist, proceeding with creation").
				WithString("media_id", media.ID).
				Log()
		default:
			return nil, NewInternalError(ctx, "write_media", "check_existing", err).
				WithMediaID(media.ID)
		}
	}

	// Content reading with logging
	tracer.Step("content_reading").
		WithString("filename", media.Filename).
		Log()

	start := time.Now()
	content, err := media.Content()
	if err != nil {
		return nil, NewMediaProcessingError(ctx, "read_content", media.Filename, err)
	}

	// Read all content into memory to compute hash and write to disk
	contentBytes, err := io.ReadAll(content)
	contentReadDuration := time.Since(start)
	if err != nil {
		return nil, NewMediaProcessingError(ctx, "read_content_bytes", media.Filename, err)
	}

	debug.FileOperation("read_content", media.Filename, int64(len(contentBytes)), contentReadDuration)

	// Compute SHA256 hash
	tracer.Step("hash_computation").
		WithInt("content_size", len(contentBytes)).
		Log()

	hash := sha256.Sum256(contentBytes)
	hashStr := fmt.Sprintf("%x", hash)

	if oldMedia != nil && hashStr == oldMedia.Hash {
		debug.BusinessLogic("media content unchanged, skipping processing").
			WithString("hash", hashStr).
			WithString("filename", media.Filename).
			Log()
		tracer.Success().
			WithString("media_id", oldMedia.ID).
			WithString("filename", oldMedia.Filename).
			WithString("hash", hashStr).
			WithBool("skipped", true).
			WithString("reason", "unchanged_content").
			Log()
		return oldMedia, nil
	}

	// Transaction with detailed step logging
	tracer.Step("transaction_start").
		WithInt("content_size", len(contentBytes)).
		WithString("hash", hashStr).
		WithBool("is_update", oldMedia != nil).
		Log()

	debug.Transaction("starting").
		WithString("operation", "write_media").
		Log()

	txStart := time.Now()
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		media.Hash = hashStr

		// Step 1: Initialize processing service
		tracer.Step("processing_init").
			WithString("filename", media.Filename).
			Log()

		processingSrv, err := NewProcessingMediaService()
		if err != nil {
			return NewMediaProcessingError(ctx, "init_processing", media.Filename, err)
		}

		// Step 2: Process media (thumbnail + EXIF)
		tracer.Step("media_processing").
			WithString("filename", media.Filename).
			WithInt("content_size", len(contentBytes)).
			Log()

		procStart := time.Now()
		r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
		processingDuration := time.Since(procStart)
		if err != nil {
			return NewMediaProcessingError(ctx, "generate_thumbnail", media.Filename, err)
		}

		debug.Processing("media_processing", media.Filename).
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

		debug.Processing("thumbnail_generated", media.Filename).
			WithInt("thumbnail_size", len(thumbnail)).
			WithInt("exif_fields", len(exif)).
			Log()

		// Step 4: Extract capture time
		if captureAt, err := media.GetCapturedTime(); err != nil {
			debug.Processing("capture_time_extraction_failed", media.Filename).
				WithString("error", err.Error()).
				Log()
		} else {
			media.CapturedAt = captureAt
			debug.Processing("capture_time_extracted", media.Filename).
				WithParam("captured_at", captureAt).
				Log()
		}

		// Step 5: Write file to disk
		tracer.Step("filesystem_write").
			WithString("filepath", media.Filepath()).
			Log()

		fsStart := time.Now()
		if err := m.fs.Write(ctx, media.Filepath(), bytes.NewReader(contentBytes)); err != nil {
			return NewFilesystemError(ctx, "write_media", "filesystem_write", media.Filepath(), err)
		}
		debug.FileOperation("write_file", media.Filepath(), int64(len(contentBytes)), time.Since(fsStart))

		// Step 6: Write to database
		tracer.Step("database_write").
			WithString("table", "media").
			Log()

		dbStart := time.Now()
		if err := writer.WriteMedia(ctx, media); err != nil {
			return NewDatabaseWriteError(ctx, "write_media", err).
				WithMediaID(media.ID).
				WithFilename(media.Filename)
		}
		tracer.Performance("database_write", time.Since(dbStart))

		return nil
	})
	transactionDuration := time.Since(txStart)

	tracer.Performance("transaction_duration", transactionDuration)

	if err != nil {
		return nil, NewInternalError(ctx, "write_media", "transaction", err).
			WithMediaID(media.ID).
			WithFilename(media.Filename)
	}

	debug.Transaction("completed").
		WithString("operation", "write_media").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		WithString("hash", hashStr).
		WithInt("content_size", len(contentBytes)).
		WithInt("thumbnail_size", len(media.Thumbnail)).
		WithInt("exif_fields", len(media.Exif)).
		WithParam("captured_at", media.CapturedAt).
		Log()

	return &media, nil
}

// UpdateMedia updates an existing media item using an entity.Media
func (m *MediaService) UpdateMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	debug := m.debug.WithContext(ctx)
	tracer := debug.StartOperation("update_media").
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		Build()

	// Clear the content function to avoid writing file content during update
	media.Content = nil

	debug.BusinessLogic("metadata-only update, content function cleared").
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		Log()

	// Database transaction with debug timing
	tracer.Step("database_update").
		WithString("table", "media").
		Log()

	debug.Transaction("starting").
		WithString("operation", "update_media").
		Log()

	start := time.Now()
	err := m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteMedia(ctx, media)
	})
	transactionDuration := time.Since(start)

	tracer.Performance("transaction_duration", transactionDuration)

	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "update_media", err).
			WithMediaID(media.ID).
			WithFilename(media.Filename)
	}

	debug.Transaction("completed").
		WithString("operation", "update_media").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		Log()

	return &media, nil
}

// DeleteMedia deletes a media item by ID
func (m *MediaService) DeleteMedia(ctx context.Context, id string) error {
	debug := m.debug.WithContext(ctx)
	tracer := debug.StartOperation("delete_media").
		WithString("media_id", id).
		Build()

	// Check if media exists
	tracer.Step("validate_exists").
		WithString("media_id", id).
		Log()

	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return NewInternalError(ctx, "delete_media", "validate_exists", err).
			WithMediaID(id)
	}

	debug.BusinessLogic("media found, proceeding with deletion").
		WithString("media_id", media.ID).
		WithString("filename", media.Filename).
		WithString("filepath", media.Filepath()).
		Log()

	// Delete the media from the datastore using a write transaction
	tracer.Step("transaction_start").
		WithString("filepath", media.Filepath()).
		WithParam("operations", []string{"filesystem_delete", "database_delete"}).
		Log()

	debug.Transaction("starting").
		WithString("operation", "delete_media").
		Log()

	start := time.Now()
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Remove the file from album folders
		tracer.Step("filesystem_delete").
			WithString("filepath", media.Filepath()).
			Log()

		fsStart := time.Now()
		if err := m.fs.DeleteMedia(ctx, media.Filepath()); err != nil {
			return NewFilesystemError(ctx, "delete_media", "filesystem_delete", media.Filepath(), err)
		}
		debug.FileOperation("delete_media", media.Filepath(), 0, time.Since(fsStart))

		// Delete the media from the database
		tracer.Step("database_delete").
			WithString("table", "media").
			WithString("media_id", id).
			Log()

		dbStart := time.Now()
		if err := writer.DeleteMedia(ctx, id); err != nil {
			return NewDatabaseWriteError(ctx, "delete_media", err).
				WithMediaID(id)
		}
		tracer.Performance("database_delete", time.Since(dbStart))

		return nil
	})
	transactionDuration := time.Since(start)

	tracer.Performance("transaction_duration", transactionDuration)

	if err != nil {
		return NewInternalError(ctx, "delete_media", "transaction", err).
			WithMediaID(id).
			WithFilepath(media.Filepath())
	}

	debug.Transaction("completed").
		WithString("operation", "delete_media").
		WithBool("success", true).
		Log()

	tracer.Success().
		WithString("media_id", id).
		WithString("filepath", media.Filepath()).
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
