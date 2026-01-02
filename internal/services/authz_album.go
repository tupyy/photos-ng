// Package services provides authorization-wrapped album service implementations.
// This file contains the AuthzAlbumService which wraps AlbumService with authorization checks.
package services

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

// AuthzAlbumService wraps AlbumService with authorization checks.
// Operations check permissions on album resources using the Authz service.
type AuthzAlbumService struct {
	albumSrv *AlbumService
	authzSrv Authz
	logger   *logger.StructuredLogger
}

// NewAuthzAlbumService creates a new authorization-wrapped album service.
// It requires an authorization service, database datastore, and filesystem datastore.
func NewAuthzAlbumService(authzSrv Authz, albumSrv *AlbumService) *AuthzAlbumService {
	return &AuthzAlbumService{
		albumSrv: albumSrv,
		authzSrv: authzSrv,
		logger:   logger.New("authz_album_service"),
	}
}

// List returns albums that the authenticated user has view permission on.
// Filters albums based on ListResources authorization check.
func (s *AuthzAlbumService) List(ctx context.Context, opts *ListOptions) ([]entity.Album, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_list_albums").Build()

	user := user.MustFromContext(ctx)

	logger.Step("list_allowed_resources").Log()

	allowedIds, err := s.authzSrv.ListResources(ctx, "", user, entity.ViewPermission, entity.AlbumResource)
	if err != nil {
		return nil, NewInternalError(ctx, "authz_list_albums", "list_allowed_resources", err)
	}

	listOptions := NewListOptionsWithOptions(opts.ToOption(), SetAllowedIDs(allowedIds))

	return s.albumSrv.List(ctx, listOptions)
}

// Count returns the total number of albums.
// Note: This doesn't require authorization filtering.
func (s *AuthzAlbumService) Count(ctx context.Context, hasParent bool) (int, error) {
	// Count doesn't require authorization filtering
	return s.albumSrv.Count(ctx, hasParent)
}

// Get retrieves a specific album by ID.
// Requires entity.ViewPermission on the album resource.
func (s *AuthzAlbumService) Get(ctx context.Context, id string) (*entity.Album, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_get_album").
		WithString(AlbumID, id).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_view_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewAlbumResource(id), entity.ViewPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_get_album", entity.NewAlbumResource(id), entity.ViewPermission)
	}

	logger.Step("authorization granted").WithString("method", "get_album").Log()
	return s.albumSrv.Get(ctx, id)
}

// Create creates a new album.
// Requires entity.CreatePermission on entity.LocalDatastore.
// Creates authorization relationships: datastore, owner, and optionally parent.
func (s *AuthzAlbumService) Create(ctx context.Context, album entity.Album) (*entity.Album, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_create_album").
		WithString(AlbumID, album.ID).
		WithString(AlbumPath, album.Path).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_create_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewDatastoreResource(entity.LocalDatastore), entity.CreatePermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_create_album", entity.NewDatastoreResource(entity.LocalDatastore), entity.CreatePermission)
	}

	// Write authorization relationships first
	logger.Step("write_authorization_relationships").Log()
	relationships := []entity.Relationship{
		entity.NewRelationship(
			entity.NewDatastoreSubject(entity.LocalDatastore),
			entity.NewAlbumResource(album.ID),
			entity.DatastoreRelationship,
		),
		entity.NewRelationship(
			entity.NewUserSubject(user.Username),
			entity.NewAlbumResource(album.ID),
			entity.OwnerRelationship,
		),
	}

	// TODO: Check permission on the parent
	if album.ParentId != nil {
		relationships = append(relationships, entity.NewRelationship(
			entity.NewAlbumSubject(*album.ParentId),
			entity.NewAlbumResource(album.ID),
			entity.ParentRelationship,
		))
	}

	err = s.authzSrv.WriteRelationships(ctx, relationships...)
	if err != nil {
		return nil, NewDatabaseWriteError(ctx, "authz_create_album", err).WithAlbumID(album.ID)
	}

	createdAlbum, err := s.albumSrv.Create(ctx, album)
	if err != nil {
		logger.Step("create_failed_rolling_back_relationships").Log()
		if delErr := s.authzSrv.DeleteRelationships(ctx, entity.NewAlbumResource(album.ID)); delErr != nil {
			logger.Step("rollback_failed").WithString("error", delErr.Error()).Log()
		}
		return nil, err
	}

	logger.Success().Log()
	return createdAlbum, nil
}

// Update updates an existing album.
// Requires entity.EditPermission on the album resource.
func (s *AuthzAlbumService) Update(ctx context.Context, album entity.Album) (*entity.Album, error) {
	logger := s.logger.WithContext(ctx).Debug("authz_update_album").
		WithString(AlbumID, album.ID).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_edit_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewAlbumResource(album.ID), entity.EditPermission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, NewForbiddenAccessError(ctx, "authz_update_album", entity.NewAlbumResource(album.ID), entity.EditPermission)
	}

	return s.albumSrv.Update(ctx, album)
}

// Delete deletes an album by ID.
// Requires entity.DeletePermission on the album resource.
// Also deletes associated authorization relationships.
func (s *AuthzAlbumService) Delete(ctx context.Context, id string) error {
	logger := s.logger.WithContext(ctx).Debug("authz_delete_album").
		WithString(AlbumID, id).
		Build()

	user := user.MustFromContext(ctx)

	logger.Step("check_delete_permission").Log()
	hasPermission, err := s.authzSrv.HasPermission(ctx, user, entity.NewAlbumResource(id), entity.DeletePermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return NewForbiddenAccessError(ctx, "authz_delete_album", entity.NewAlbumResource(id), entity.DeletePermission)
	}

	if err := s.albumSrv.Delete(ctx, id); err != nil {
		return err
	}

	// If successful, delete authorization relationships
	logger.Step("delete_authorization_relationships").Log()
	err = s.authzSrv.DeleteRelationships(ctx, entity.NewAlbumResource(id))
	if err != nil {
		logger.Step("failed_to_delete_relationships").WithString("error", err.Error()).Log()
		// Note: Album is already deleted, but relationships cleanup failed
		// This is logged but we don't fail the operation
	}

	logger.Success().Log()
	return nil
}
