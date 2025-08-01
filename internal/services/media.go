package services

import (
	"context"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// MediaService provides business logic for media operations
type MediaService struct {
	dt *pg.Datastore
}

// NewMediaService creates a new instance of MediaService with the provided datastore
func NewMediaService(dt *pg.Datastore) *MediaService {
	return &MediaService{dt: dt}
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
	media, err := m.dt.QueryMedia(ctx, pg.FilterById(id), pg.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(media) == 0 {
		return nil, &ErrResourceNotFound{Resource: "media", ID: id}
	}

	return &media[0], nil
}

// UpdateMedia updates an existing media item using the v1 API request
func (m *MediaService) UpdateMedia(ctx context.Context, id string, request v1.UpdateMediaRequest) (*entity.Media, error) {
	// Get the existing media
	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates if provided
	if request.CapturedAt != nil {
		media.CapturedAt = request.CapturedAt.Time
	}
	if request.Exif != nil {
		if media.Exif == nil {
			media.Exif = make(map[string]string)
		}
		for _, exif := range *request.Exif {
			media.Exif[exif.Key] = exif.Value
		}
	}

	// TODO: Implement media update in datastore
	// For now, return the updated media (stub implementation)
	return media, nil
}

// DeleteMedia deletes a media item by ID
func (m *MediaService) DeleteMedia(ctx context.Context, id string) error {
	// Check if media exists
	_, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Implement media deletion in datastore
	// For now, this is a stub implementation
	return nil
}

// GetMediaContent retrieves the file content for a media item
func (m *MediaService) GetMediaContent(ctx context.Context, id string) (*entity.Media, []byte, error) {
	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Implement actual file reading
	// For now, return placeholder content
	content := []byte("placeholder media content")

	return media, content, nil
}

// GetMediaThumbnail retrieves the thumbnail for a media item
func (m *MediaService) GetMediaThumbnail(ctx context.Context, id string) (*entity.Media, []byte, error) {
	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if !media.HasThumbnail() {
		return nil, nil, &ErrResourceNotFound{Resource: "thumbnail", ID: id}
	}

	return media, media.Thumbnail, nil
}
