package services

import (
	"context"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// MediaFilter represents filtering criteria for media queries
type MediaFilter struct {
	Limit      int
	Offset     int
	AlbumID    *string
	MediaType  *string
	StartDate  *time.Time
	EndDate    *time.Time
	SortBy     string
	Descending bool
}

// QueriesFn returns a slice of query options based on the media filter criteria
func (mf *MediaFilter) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Add pagination
	if mf.Limit > 0 {
		qf = append(qf, pg.Limit(mf.Limit))
	}
	if mf.Offset > 0 {
		qf = append(qf, pg.Offset(mf.Offset))
	}

	// Add album filter
	if mf.AlbumID != nil {
		qf = append(qf, pg.FilterByColumnName("album_id", *mf.AlbumID))
	}

	// Add media type filter
	if mf.MediaType != nil {
		qf = append(qf, pg.FilterByColumnName("media_type", *mf.MediaType))
	}

	// Add date range filters
	// TODO: Implement date range filtering in query options
	// For now, we'll filter in the service layer after querying

	// Add sorting
	if mf.SortBy != "" {
		qf = append(qf, pg.SortByColumn(mf.SortBy, mf.Descending))
	} else {
		// Default sort by captured_at descending
		qf = append(qf, pg.SortByColumn("captured_at", true))
	}

	return qf
}

// MediaService provides business logic for media operations
type MediaService struct {
	dt *pg.Datastore
}

// NewMediaService creates a new instance of MediaService with the provided datastore
func NewMediaService(dt *pg.Datastore) *MediaService {
	return &MediaService{dt: dt}
}

// GetMedia retrieves a list of media items based on the provided filter criteria
func (m *MediaService) GetMedia(ctx context.Context, filter *MediaFilter) ([]entity.Media, error) {
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

// UpdateMedia updates an existing media item
func (m *MediaService) UpdateMedia(ctx context.Context, id string, updateFn func(*entity.Media) error) (*entity.Media, error) {
	// Get the existing media
	media, err := m.GetMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply the update function
	if err := updateFn(media); err != nil {
		return nil, err
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
