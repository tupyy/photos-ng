package services

import (
	"context"
	"fmt"
	"path"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// AlbumService provides business logic for album operations
type AlbumService struct {
	dt *pg.Datastore
	fs *fs.Datastore
}

// NewAlbumService creates a new instance of AlbumService with the provided datastores
func NewAlbumService(dt *pg.Datastore, fsDatastore *fs.Datastore) *AlbumService {
	return &AlbumService{dt: dt, fs: fsDatastore}
}

// GetAlbums retrieves a list of albums based on the provided filter criteria
// Handles pagination at the application level to avoid issues with JOIN queries
func (a *AlbumService) GetAlbums(ctx context.Context, opts *AlbumOptions) ([]entity.Album, error) {
	// Get all albums matching the filter criteria (without pagination)
	allAlbums, err := a.dt.QueryAlbums(ctx, opts.QueriesFn()...)
	if err != nil {
		return nil, err
	}

	// Apply pagination at the application level
	start := opts.Offset
	end := opts.Offset + opts.Limit

	// Handle bounds
	if start >= len(allAlbums) {
		return []entity.Album{}, nil
	}
	if end > len(allAlbums) || opts.Limit <= 0 {
		end = len(allAlbums)
	}

	// Return the paginated slice
	return allAlbums[start:end], nil
}

// GetAlbum retrieves a specific album by its ID
func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
	albums, err := a.dt.QueryAlbums(ctx, pg.FilterByAlbumId(id))
	if err != nil {
		return nil, err
	}

	if len(albums) == 0 {
		return nil, NewErrAlbumNotFound(id)
	}

	return &albums[0], nil
}

// CreateAlbum creates a new album using an entity.Album
func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	// Check if the album already exists
	isAlbumExists := true
	if _, err := a.GetAlbum(ctx, album.ID); err != nil && IsErrResourceNotFound(err) {
		isAlbumExists = false
	}

	// Get parent if parentID exists and recompute path and id of the new album.
	if album.ParentId != nil {
		parent, err := a.GetAlbum(ctx, *album.ParentId)
		if err != nil {
			if IsErrResourceNotFound(err) {
				return nil, NewErrResourceNotFound(fmt.Errorf("album %s parent does not exists", album.ID))
			}
			return nil, err
		}
		album.Path = path.Join(parent.Path, album.Path)
		album.ID = entity.GenerateId(album.Path)
	}

	// Create the album in the datastore using a write transaction
	err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
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

// CreateAlbum creates a new album using an entity.Album
func (a *AlbumService) UpdateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
	existingAlbum, err := a.GetAlbum(ctx, album.ID)
	if err != nil {
		return nil, err
	}

	existingAlbum.Description = album.Description

	// if thumbnail is present, check if the media belongs to the album
	if album.Thumbnail != nil {
		media, err := a.dt.QueryMedia(ctx, pg.FilterByMediaId(*album.Thumbnail), pg.FilterByColumnName("album_id", existingAlbum.ID), pg.Limit(1))
		if err != nil {
			return nil, err
		}

		if len(media) == 0 {
			return nil, NewErrUpdateAlbum(fmt.Sprintf("thumbnail %s does not exists in the album", *album.Thumbnail))
		}

		existingAlbum.Thumbnail = album.Thumbnail
	}

	// Write the album in the datastore using a write transaction
	err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		return writer.WriteAlbum(ctx, *existingAlbum)
	})
	if err != nil {
		return nil, err
	}

	return existingAlbum, nil
}

// DeleteAlbum deletes an album by ID
func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
	// Check if album exists
	if _, err := a.GetAlbum(ctx, id); err != nil {
		return err
	}

	// Delete the album from the datastore using a write transaction
	err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		// Delete the album folder from the file system
		if err := a.fs.DeleteFolder(ctx, id); err != nil {
			return err
		}

		// Delete the album from the database
		return writer.DeleteAlbum(ctx, id)

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
