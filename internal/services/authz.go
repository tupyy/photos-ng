package services

import (
	"context"
	"slices"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/authz"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type Authz interface {
	// WriteRelationships writes authorization relationships to the datastore
	WriteRelationships(ctx context.Context, relationships ...entity.Relationship) error

	// HasPermission checks if a user has a specific permission on a resource
	HasPermission(ctx context.Context, user entity.User, resource entity.Resource, permission entity.Permission) (bool, error)

	// GetPermissions returns all permissions a user has on a list of resources
	GetPermissions(ctx context.Context, zedToken string, user entity.User, resources []entity.Resource) (map[entity.Resource][]entity.Permission, error)

	// ListResources returns all resource IDs of a given type that a user can access with a specific permission
	ListResources(ctx context.Context, zedToken string, user entity.User, permission entity.Permission, resourceKind entity.ResourceKind) ([]string, error)

	// DeleteRelationships deletes all relationships for a resource
	DeleteRelationships(ctx context.Context, resource entity.Resource) error
}

type AuthzService struct {
	authzStore *authz.Datastore
	pg         *pg.Datastore
}

func NewAuthzService(authzS *authz.Datastore, pg *pg.Datastore) *AuthzService {
	return &AuthzService{
		authzStore: authzS,
		pg:         pg,
	}
}

func (s *AuthzService) HasPermission(ctx context.Context, user entity.User, resource entity.Resource, permission entity.Permission) (bool, error) {
	hasPermission := false
	if err := s.pg.ExecWithSharedLock(ctx, func(ctx context.Context) error {
		token, err := s.pg.ReadToken(ctx)
		if err != nil {
			return err
		}

		permissions, err := s.authzStore.GetPermissions(ctx, token, user.ID, []entity.Resource{resource})
		if err != nil {
			return err
		}

		userPermissions, exists := permissions[resource]
		if !exists {
			return nil
		}

		hasPermission = slices.ContainsFunc(userPermissions, func(p entity.Permission) bool {
			return p == permission
		})
		return nil
	}); err != nil {
		return hasPermission, err
	}
	return hasPermission, nil
}

func (s *AuthzService) GetPermissions(ctx context.Context, zedToken string, user entity.User, resources []entity.Resource) (map[entity.Resource][]entity.Permission, error) {
	var permissions map[entity.Resource][]entity.Permission
	err := s.pg.ExecWithSharedLock(ctx, func(ctx context.Context) error {
		token, err := s.pg.ReadToken(ctx)
		if err != nil {
			return err
		}

		permissions, err = s.authzStore.GetPermissions(ctx, token, user.ID, resources)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s *AuthzService) ListResources(ctx context.Context, zedToken string, user entity.User, permission entity.Permission, resourceKind entity.ResourceKind) ([]string, error) {
	var allowedIds []string
	err := s.pg.ExecWithSharedLock(ctx, func(ctx context.Context) error {
		token, err := s.pg.ReadToken(ctx)
		if err != nil {
			return err
		}

		resource := entity.Resource{
			Kind: resourceKind,
			// ID is empty for listing all resources of a type
		}
		allowedIds, err = s.authzStore.ListResources(ctx, token, user.ID, permission, resource)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return allowedIds, nil
}

func (s *AuthzService) DeleteRelationships(ctx context.Context, resource entity.Resource) error {
	return s.pg.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		writer.AcquireGlobalLock(ctx)

		token, err := s.authzStore.DeleteRelationships(ctx, resource)
		if err != nil {
			return err
		}

		if err := writer.WriteToken(ctx, token); err != nil {
			return err
		}

		return nil
	})
}

func (s *AuthzService) WriteRelationships(ctx context.Context, relationships ...entity.Relationship) error {
	relationshipFns := make([]authz.RelationshipFn, 0, len(relationships))
	for _, rel := range relationships {
		relationshipFns = append(relationshipFns, authz.WithRelationship(rel.Subject, rel.Resource, rel.Kind))
	}

	return s.pg.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
		writer.AcquireGlobalLock(ctx)

		token, err := s.authzStore.WriteRelationships(ctx, relationshipFns...)
		if err != nil {
			return err
		}

		if err := writer.WriteToken(ctx, token); err != nil {
			return err
		}

		return nil
	})
}
