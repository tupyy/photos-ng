package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"path"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"go.uber.org/zap"
)

// MediaService provides business logic for media operations
type MediaService struct {
	dt *pg.Datastore
	fs *fs.Datastore
}

// NewMediaService creates a new instance of MediaService with the provided datastores
func NewMediaService(dt *pg.Datastore, fsDatastore *fs.Datastore) *MediaService {
	return &MediaService{dt: dt, fs: fsDatastore}
}

// GetMedia retrieves a list of media items based on the provided filter criteria
func (m *MediaService) GetMedia(ctx context.Context, filter *MediaOptions) ([]entity.Media, error) {
	media, err := m.dt.QueryMedia(ctx, filter.QueriesFn()...)
	if err != nil {
		return nil, err
	}

	// Apply date filtering in-memory for now
	// TODO: Move this to database-level filtering
	if filter.StartDate != nil || filter.EndDate != nil {
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
		media = filteredMedia
	}

	return media, nil
}

// GetMediaByID retrieves a specific media item by its ID
func (m *MediaService) GetMediaByID(ctx context.Context, id string) (*entity.Media, error) {
	if id == "" {
		return nil, NewValidationError(ctx, "get_media", "invalid_input").
			WithContext("validation_error", "empty_media_id")
	}

	media, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(id), pg.Limit(1))
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "get_media", err).
			WithMediaID(id).
			AtStep("query_media")
	}

	if len(media) == 0 {
		return nil, NewMediaNotFoundError(ctx, id)
	}

	processedMedia := media[0]

	// Populate the content function using the filesystem datastore
	processedMedia.Content = m.fs.Read(ctx, processedMedia.Filepath())

	return &processedMedia, nil
}

// WriteMedia creates or updates a media item and writes its content to disk
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	// Check if the media already exists
	oldMedia, err := m.GetMediaByID(ctx, media.ID)
	if err != nil {
		var notFoundErr *NotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, NewInternalError(ctx, "write_media", "check_existing", err).
				WithMediaID(media.ID)
		}
	}

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
	hash := sha256.Sum256(contentBytes)
	hashStr := fmt.Sprintf("%x", hash)

	if oldMedia != nil && hashStr == oldMedia.Hash {
		zap.S().Debugw("update media skipped.same hash", "hash", oldMedia.Hash)
		return oldMedia, nil
	}

	// Write the media to the datastore using a write transaction
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		media.Hash = hashStr

		// process the photo
		processingSrv, err := NewProcessingMediaService()
		if err != nil {
			return NewMediaProcessingError(ctx, "init_processing", media.Filename, err)
		}

		r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
		if err != nil {
			return NewMediaProcessingError(ctx, "generate_thumbnail", media.Filename, err)
		}

		thumbnail, err := io.ReadAll(r)
		if err != nil {
			return NewMediaProcessingError(ctx, "read_thumbnail", media.Filename, err)
		}
		media.Thumbnail = thumbnail
		media.Exif = exif

		if captureAt, err := media.GetCapturedTime(); err != nil {
			zap.S().Warnw("failed to get captured at timestamp", "error", err, "filename", media.Filepath())
		} else {
			media.CapturedAt = captureAt
		}

		// Write the file to disk using the media's filepath
		if err := m.fs.Write(ctx, media.Filepath(), bytes.NewReader(contentBytes)); err != nil {
			return NewFilesystemError(ctx, "write_media", "filesystem_write", media.Filepath(), err)
		}
		// Write the media to database (with computed hash)
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

	return &media, nil
}

// UpdateMedia updates an existing media item using an entity.Media
func (m *MediaService) UpdateMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	// Clear the content function to avoid writing file content during update
	media.Content = nil

	// Update the media metadata in the datastore using a write transaction
	err := m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteMedia(ctx, media)
	})
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "update_media", err).
			WithMediaID(media.ID).
			WithFilename(media.Filename)
	}

	return &media, nil
}

// DeleteMedia deletes a media item by ID
func (m *MediaService) DeleteMedia(ctx context.Context, id string) error {
	// Check if media exists
	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return NewInternalError(ctx, "delete_media", "validate_exists", err).
			WithMediaID(id)
	}

	// Delete the media from the datastore using a write transaction
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Remove the file from album folders
		if err := m.fs.DeleteMedia(ctx, media.Filepath()); err != nil {
			return NewFilesystemError(ctx, "delete_media", "filesystem_delete", media.Filepath(), err)
		}
		// Delete the media from the database
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

	return nil
}

func (m *MediaService) GetContentFn(ctx context.Context, media entity.Media) entity.MediaContentFn {
	// Construct the full file path from media filename and album path
	filepath := path.Join(media.Album.Path, media.Filename)
	return m.fs.Read(ctx, filepath)
}
