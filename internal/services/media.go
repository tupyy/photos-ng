package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

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
	media, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(id), pg.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(media) == 0 {
		return nil, NewErrMediaNotFound(id)
	}

	processedMedia := media[0]

	// Populate the content function using the filesystem datastore
	processedMedia.Content = m.fs.Read(ctx, processedMedia.Filepath())

	return &processedMedia, nil
}

// WriteMedia creates or updates a media item and writes its content to disk
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	// Check if the media already exists
	isMediaExists := true
	_, err := m.GetMediaByID(ctx, media.ID)
	if err != nil && IsErrResourceNotFound(err) {
		isMediaExists = false
	}

	// Write the media to the datastore using a write transaction
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		if isMediaExists {
			return writer.WriteMedia(ctx, media)
		}

		content, err := media.Content()
		if err != nil {
			return err
		}

		// Read all content into memory to compute hash and write to disk
		contentBytes, err := io.ReadAll(content)
		if err != nil {
			return fmt.Errorf("failed to read media content: %w", err)
		}

		// Compute SHA256 hash
		hash := sha256.Sum256(contentBytes)
		media.Hash = fmt.Sprintf("%x", hash)

		// process the photo
		processingSrv, err := NewProcessingMediaService()
		if err != nil {
			return err
		}

		r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
		if err != nil {
			return err
		}

		thumbnail, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read thumbnail reader from processing service: %w", err)
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
			return err
		}
		// Write the media to database (with computed hash)
		return writer.WriteMedia(ctx, media)
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &media, nil
}

// DeleteMedia deletes a media item by ID
func (m *MediaService) DeleteMedia(ctx context.Context, id string) error {
	// Check if media exists
	_, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the media from the datastore using a write transaction
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Delete the media from the database
		return writer.DeleteMedia(ctx, id)
	})
	if err != nil {
		return err
	}

	return nil
}
