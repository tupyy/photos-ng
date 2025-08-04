package services

import (
	"context"
	"errors"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
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

	return &media[0], nil
}

// WriteMedia creates or updates a media item and writes its content to disk
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	// Check if the media already exists
	isMediaExists := true
	_, err := m.GetMediaByID(ctx, media.ID)
	if err == nil && errors.Is(err, NewErrMediaNotFound(media.ID)) {
		isMediaExists = false
	}

	// Write the media to the datastore using a write transaction
	err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Write the media to database
		if err := writer.WriteMedia(ctx, media); err != nil {
			return err
		}

		// If it's a new media and has content, write the file to disk
		if !isMediaExists {
			content, err := media.Content()
			if err != nil {
				return err
			}

			// Write the file to disk using the media's filepath
			if err := m.fs.Write(ctx, media.Filepath(), content); err != nil {
				return err
			}
		}

		return nil
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
