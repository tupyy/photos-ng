package services

import (
	"context"
	"errors"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// AlbumService provides business logic for album operations
type AlbumService struct {
	dt *pg.Datastore
	fs *fs.FsDatastore
}

// NewAlbumService creates a new instance of AlbumService with the provided datastores
func NewAlbumService(dt *pg.Datastore, fsDatastore *fs.FsDatastore) *AlbumService {
	return &AlbumService{dt: dt, fs: fsDatastore}
}

// GetAlbums retrieves a list of albums based on the provided filter criteria
func (a *AlbumService) GetAlbums(ctx context.Context, opts *AlbumOptions) ([]entity.Album, error) {
	return a.dt.QueryAlbums(ctx, opts.QueriesFn()...)
}

// GetAlbum retrieves a specific album by its ID
func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
	albums, err := a.dt.QueryAlbums(ctx, pg.FilterById(id), pg.Limit(1))
	if err != nil {
		return nil, err
	}

	if len(albums) == 0 {
		return nil, NewErrAlbumNotFound(id)
	}

	return &albums[0], nil
}

// CreateAlbum creates a new album using an entity.Album
func (a *AlbumService) WriteAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	// Check if the album already exists
	isAlbumExists := false
	_, err := a.GetAlbum(ctx, album.ID)
	if err == nil && errors.Is(err, NewErrAlbumNotFound(album.ID)) {
		isAlbumExists = true
	}

	// Create the album in the datastore using a write transaction
	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Write the album to database
		if err := writer.WriteAlbum(ctx, album); err != nil {
			return err
		}

		// If it's a new album, create the folder on disk
		if !isAlbumExists {
			if err := a.fs.CreateFolder(ctx, album.Path); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &album, nil
}

// UpdateAlbum updates an existing album using the v1 API request
func (a *AlbumService) UpdateAlbum(ctx context.Context, id string, request v1.UpdateAlbumRequest) (*entity.Album, error) {
	// Get the existing album
	album, err := a.GetAlbum(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates if provided
	if request.Name != nil {
		album.Path = *request.Name // Using name as path for now
	}
	// TODO: Add description field to entity.Album if needed
	// if request.Description != nil {
	//     album.Description = *request.Description
	// }

	// TODO: Implement album update in datastore
	// For now, return the updated album (stub implementation)
	return album, nil
}

// DeleteAlbum deletes an album by ID
func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
	// Check if album exists
	_, err := a.GetAlbum(ctx, id)
	if err != nil {
		return err
	}

	// Delete the album from the datastore using a write transaction
	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Delete the album from the database
		if err := writer.DeleteAlbum(ctx, id); err != nil {
			return err
		}

		// Delete the album folder from the file system
		return a.fs.DeleteFolder(ctx, id)
	})
	if err != nil {
		return err
	}

	return nil
}

// SyncAlbum synchronizes an album with the file system
func (a *AlbumService) SyncAlbum(ctx context.Context, id string) (int, error) {
	// Check if album exists
	album, err := a.GetAlbum(ctx, id)
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
