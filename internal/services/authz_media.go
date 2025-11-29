// Package services provides authorization-wrapped media service implementations.
// This file contains the AuthzMediaService which wraps MediaService with authorization checks.
package services

import (
	"context"
	"slices"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

const (
	maxPageSize = 100
)

// AuthzMediaService wraps MediaService with authorization checks.
// Operations check permissions on media resources using the Authz service.
type AuthzMediaService struct {
	mediaSrv *MediaService
	authzSrv Authz
	logger   *logger.StructuredLogger
}

// NewAuthzMediaService creates a new authorization-wrapped media service.
// It requires an authorization service, database datastore, and filesystem datastore.
func NewAuthzMediaService(authzSrv Authz, dt *pg.Datastore, fsDatastore *fs.Datastore) *AuthzMediaService {
	return &AuthzMediaService{
		mediaSrv: NewMediaService(dt, fsDatastore),
		authzSrv: authzSrv,
		logger:   logger.New("authz_media_service"),
	}
}

// List retrieves media items based on filter criteria with authorization filtering.
// Uses a loop-until-full approach: fetches batches from the database, filters by permission,
// and continues until the requested page size is reached or no more data is available.
func (s *AuthzMediaService) List(ctx context.Context, filter *MediaOptions) ([]entity.Media, *PaginationCursor, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_list_media").
		WithStringPtr(AlbumID, filter.AlbumID).
		WithInt("requested_limit", filter.MediaLimit).
		Build()

	user := user.MustFromContext(ctx)
	requestedLimit := filter.MediaLimit
	if requestedLimit <= 0 {
		requestedLimit = maxPageSize
	}

	// Use a larger batch size to reduce round trips
	filter.MediaLimit = requestedLimit * 2

	var results []entity.Media
	var lastCursor *PaginationCursor
	totalFetched := 0

	for len(results) < requestedLimit {
		logger.Step("fetch_batch").
			WithInt("batch_size", filter.MediaLimit).
			WithInt("results_so_far", len(results)).
			Log()

		batch, nextCursor, err := s.mediaSrv.List(ctx, filter)
		if err != nil {
			return nil, nil, err
		}

		totalFetched += len(batch)

		if len(batch) == 0 {
			break
		}

		// Build resources for batch permission check
		resources := make([]entity.Resource, len(batch))
		for i, media := range batch {
			resources[i] = entity.NewMediaResource(media.ID)
		}

		// Single call to check all permissions
		permissions, err := s.authzSrv.GetPermissions(ctx, "", user, resources)
		if err != nil {
			return nil, nil, err
		}

		// Filter based on permissions
		for _, media := range batch {
			resource := entity.NewMediaResource(media.ID)
			perms, hasPerms := permissions[resource]
			if hasPerms && slices.Contains(perms, entity.ViewPermission) {
				results = append(results, media)
			}

			// it does not matter if the user has or not permission on this resource
			// by advancing the cursor on jump over the media that the user cannot access
			lastCursor = &PaginationCursor{
				CapturedAt: media.CapturedAt,
				ID:         media.ID,
			}

			if len(results) >= requestedLimit {
				break
			}
		}

		if nextCursor == nil {
			break
		}
		filter.Cursor = nextCursor
	}

	// Determine the next cursor for the response
	// If we filled the page, use the last item's cursor for next page
	var responseCursor *PaginationCursor
	if len(results) >= requestedLimit && lastCursor != nil {
		responseCursor = lastCursor
	}

	logger.Success().
		WithInt("returned", len(results)).
		WithInt("total_fetched", totalFetched).
		WithInt("total_filtered", totalFetched-len(results)).
		WithBool("has_next", responseCursor != nil).
		Log()

	return results, responseCursor, nil
}

// Get retrieves a specific media item by ID.
// Requires entity.ViewPermission on the media resource.
func (s *AuthzMediaService) Get(ctx context.Context, id string) (*entity.Media, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_get_media").
		WithString(MediaID, id).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_view_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewMediaResource(id), entity.ViewPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_get_media", entity.NewMediaResource(id), entity.ViewPermission)
	}

	logger.Step("authorization granted").WithString("method", "get_media").Log()
	return s.mediaSrv.Get(ctx, id)
}

// WriteMedia creates or updates a media item.
// Requires entity.CreatePermission on the parent album resource.
// Creates authorization relationships: album parent and user owner.
func (s *AuthzMediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_write_media").
		WithString(MediaID, media.ID).
		WithString(AlbumID, media.Album.ID).
		WithString("filename", media.Filename).
		Build()

	user := user.MustFromContext(ctx)

	// Check create permission on the parent album
	logger.Step("check_create_permission_on_album").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewAlbumResource(media.Album.ID), entity.CreatePermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_write_media", entity.NewAlbumResource(media.Album.ID), entity.CreatePermission)
	}

	// Write authorization relationships first
	logger.Step("write_authorization_relationships").Log()
	relationships := []entity.Relationship{
		entity.NewRelationship(
			entity.NewAlbumSubject(media.Album.ID),
			entity.NewMediaResource(media.ID),
			entity.ParentRelationship,
		),
		entity.NewRelationship(
			entity.NewUserSubject(user.Username),
			entity.NewMediaResource(media.ID),
			entity.OwnerRelationship,
		),
	}

	err = s.authzSrv.WriteRelationships(ctx, relationships...)
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "authz_write_media", err).
			WithContext(MediaID, media.ID)
	}

	createdMedia, err := s.mediaSrv.WriteMedia(ctx, media)
	if err != nil {
		logger.Step("write_failed_rolling_back_relationships").Log()
		if delErr := s.authzSrv.DeleteRelationships(ctx, entity.NewMediaResource(media.ID)); delErr != nil {
			logger.Step("rollback_failed").WithString("error", delErr.Error()).Log()
		}
		return nil, err
	}

	logger.Success().Log()
	return createdMedia, nil
}

// Update updates an existing media item.
// Requires entity.EditPermission on the media resource.
func (s *AuthzMediaService) Update(ctx context.Context, media entity.Media) (*entity.Media, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_update_media").
		WithString(MediaID, media.ID).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_edit_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewMediaResource(media.ID), entity.EditPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_update_media", entity.NewMediaResource(media.ID), entity.EditPermission)
	}

	logger.Step("authorization granted").WithString("method", "update_media").Log()
	return s.mediaSrv.Update(ctx, media)
}

// Delete deletes a media item by ID.
// Requires entity.DeletePermission on the media resource.
// Also deletes associated authorization relationships.
func (s *AuthzMediaService) Delete(ctx context.Context, id string) error {
	logger := s.logger.WithContext(ctx).Debug("authz_delete_media").
		WithString(MediaID, id).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_delete_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewMediaResource(id), entity.DeletePermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_delete_media", entity.NewMediaResource(id), entity.DeletePermission)
	}

	logger.Step("authorization granted").WithString("method", "delete_media").Log()

	if err := s.mediaSrv.Delete(ctx, id); err != nil {
		return err
	}

	// If successful, delete authorization relationships
	logger.Step("delete_authorization_relationships").Log()
	err = s.authzSrv.DeleteRelationships(ctx, entity.NewMediaResource(id))
	if err != nil {
		logger.Step("failed_to_delete_relationships").WithString("error", err.Error()).Log()
		// Note: Media is already deleted, but relationships cleanup failed
		// This is logged but we don't fail the operation
	}

	logger.Success().Log()
	return nil
}
