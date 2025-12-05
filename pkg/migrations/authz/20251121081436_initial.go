package authz

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

func Initial(ctx context.Context, userSrv *services.UserService, authzSrv *services.AuthzService) error {
	l := logger.New("auth_migration").Info("initial_migration").Build()

	l.Step("create_datastore_migrations").Log()
	relationships := []entity.Relationship{
		entity.NewRelationship(entity.NewRoleSubject(entity.AdminRoleName), entity.NewDatastoreResource(entity.LocalDatastore), entity.AdminRelationship),
		entity.NewRelationship(entity.NewRoleSubject(entity.ViewerRoleName), entity.NewDatastoreResource(entity.LocalDatastore), entity.ViewerRelationship),
		entity.NewRelationship(entity.NewRoleSubject(entity.EditorRoleName), entity.NewDatastoreResource(entity.LocalDatastore), entity.EditorRelationship),
		entity.NewRelationship(entity.NewRoleSubject(entity.CreatorRoleName), entity.NewDatastoreResource(entity.LocalDatastore), entity.CreatorRelationship),
	}

	if err := authzSrv.WriteRelationships(ctx, relationships...); err != nil {
		return err
	}

	return nil
}
