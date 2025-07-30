package services

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// AlbumFilter represents filtering criteria for album queries
type AlbumFilter struct {
	Limit      int
	Offset     int
	ParentID   *string
	SortBy     string
	Descending bool
}

// QueriesFn returns a slice of query options based on the album filter criteria
func (af *AlbumFilter) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Add pagination
	if af.Limit > 0 {
		qf = append(qf, pg.Limit(af.Limit))
	}
	if af.Offset > 0 {
		qf = append(qf, pg.Offset(af.Offset))
	}

	// Add parent filter
	if af.ParentID != nil {
		qf = append(qf, pg.FilterByColumnName("parent", *af.ParentID))
	}

	// Add sorting
	if af.SortBy != "" {
		qf = append(qf, pg.SortByColumn(af.SortBy, af.Descending))
	} else {
		// Default sort by created_at descending
		qf = append(qf, pg.SortByColumn("created_at", true))
	}

	return qf
}

// AlbumService provides business logic for album operations
type AlbumService struct {
	dt *pg.Datastore
}

// NewAlbumService creates a new instance of AlbumService with the provided datastore
func NewAlbumService(dt *pg.Datastore) *AlbumService {
	return &AlbumService{dt: dt}
}

// GetAlbums retrieves a list of albums based on the provided filter criteria
func (a *AlbumService) GetAlbums(ctx context.Context, filter *AlbumFilter) ([]entity.Album, error) {
	return a.dt.QueryAlbums(ctx, filter.QueriesFn()...)
}

// GetAlbumByID retrieves a specific album by its ID
func (a *AlbumService) GetAlbumByID(ctx context.Context, id string) (*entity.Album, error) {
	albums, err := a.dt.QueryAlbums(ctx, pg.FilterById(id), pg.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(albums) == 0 {
		return nil, &ErrResourceNotFound{Resource: "album", ID: id}
	}

	return &albums[0], nil
}

// CreateAlbum creates a new album
func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	// TODO: Implement album creation in datastore
	// For now, return the album as-is (stub implementation)
	return &album, nil
}

// UpdateAlbum updates an existing album
func (a *AlbumService) UpdateAlbum(ctx context.Context, id string, updateFn func(*entity.Album) error) (*entity.Album, error) {
	// Get the existing album
	album, err := a.GetAlbumByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply the update function
	if err := updateFn(album); err != nil {
		return nil, err
	}

	// TODO: Implement album update in datastore
	// For now, return the updated album (stub implementation)
	return album, nil
}

// DeleteAlbum deletes an album by ID
func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
	// Check if album exists
	_, err := a.GetAlbumByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Implement album deletion in datastore
	// For now, this is a stub implementation
	return nil
}

// SyncAlbum synchronizes an album with the file system
func (a *AlbumService) SyncAlbum(ctx context.Context, id string) (int, error) {
	// Check if album exists
	album, err := a.GetAlbumByID(ctx, id)
	if err != nil {
		return 0, err
	}

	// TODO: Implement actual sync logic
	// This would typically:
	// 1. Scan the album's file system path
	// 2. Compare with database records
	// 3. Add new media files
	// 4. Remove deleted files
	// 5. Update metadata

	_ = album        // Use the album for sync logic
	syncedItems := 0 // Placeholder

	return syncedItems, nil
}
