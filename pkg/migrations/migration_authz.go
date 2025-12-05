package migrations

import (
	"context"
	"maps"
	"slices"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/migrations/authz"
)

type MigrateFn func(ctx context.Context, userSrv *services.UserService, authzSrv *services.AuthzService) error

type AuthzMigrations struct {
	migrations map[int]MigrateFn
}

func NewAuthzMigrations() *AuthzMigrations {
	return &AuthzMigrations{
		migrations: map[int]MigrateFn{
			1: authz.Initial,
		},
	}
}

func (a *AuthzMigrations) AddMigrations(key int, fn MigrateFn) {
	a.migrations[key] = fn
}

func (a *AuthzMigrations) Migrate(ctx context.Context, userSrv *services.UserService, authzSrv *services.AuthzService) error {
	keys := slices.Sorted(maps.Keys(a.migrations))

	for _, k := range keys {
		fn := a.migrations[k]
		if err := fn(ctx, userSrv, authzSrv); err != nil {
			return err
		}
	}

	return nil
}
