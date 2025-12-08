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

	users, err := userSrv.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Role == nil {
			continue
		}
		userRelationships := []entity.Relationship{}
		l.Step("create user role relationship").
			WithString("username", user.Username).
			WithString("role", user.Role.String()).
			Log()
		switch *user.Role {
		case entity.AdminRole:
			userRelationships = append(userRelationships, entity.NewRelationship(entity.NewUserSubject(user.Username), entity.NewRoleResource(entity.AdminRoleName), entity.MemberRelationship))
		case entity.CreatorRole:
			userRelationships = append(userRelationships, entity.NewRelationship(entity.NewUserSubject(user.Username), entity.NewRoleResource(entity.CreatorRoleName), entity.MemberRelationship))
		case entity.EditorRole:
			userRelationships = append(userRelationships, entity.NewRelationship(entity.NewUserSubject(user.Username), entity.NewRoleResource(entity.EditorRoleName), entity.MemberRelationship))
		case entity.ViewerRole:
			userRelationships = append(userRelationships, entity.NewRelationship(entity.NewUserSubject(user.Username), entity.NewRoleResource(entity.ViewerRoleName), entity.MemberRelationship))
		}
		if err := authzSrv.WriteRelationships(ctx, userRelationships...); err != nil {
			return err
		}
	}

	if err := authzSrv.WriteRelationships(ctx, relationships...); err != nil {
		return err
	}

	return nil
}
