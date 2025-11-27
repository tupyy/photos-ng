package cmd

import (
	"context"

	"github.com/fatih/color"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jzelinskie/cobrautil/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/authz"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/datastore"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/migrations"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/spicedb"
)

// NewMigrateCommand creates a new cobra command for running database migrations.
func NewAuthzMigrateCommand(config *config.Config) *cobra.Command {
	var migrationPath string

	cmd := &cobra.Command{
		Use:          "migrate-authz",
		Short:        "Run spicedb migrations",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			// init datastore
			pgPool, err := datastore.NewConnPool(context.Background(), config.Database.URI)
			if err != nil {
				return err
			}
			defer pgPool.Close()

			// connect to spicedb
			spiceClient, err := spicedb.InitSpiceDBClient(config.Authorization.SpiceDBURL, config.Authorization.PresharedKey)
			if err != nil {
				return err
			}

			authServ := services.NewAuthzService(authz.NewAuthzDatastore(spiceClient), pg.NewPostgresDatastore(pgPool))

			// Run migrations
			zap.S().Info("running spicedb migrations")
			if err := migrations.NewAuthzMigrations().Migrate(context.Background(), services.NewUserService(), authServ); err != nil {
				zap.S().Error("migration failed", "error", err)
				return err
			}

			zap.S().Info("spicedb migrations completed successfully")
			return nil
		},
	}

	registerAuthzMigrateFlags(cmd, config, &migrationPath)
	return cmd
}

func registerAuthzMigrateFlags(cmd *cobra.Command, config *config.Config, migrationPath *string) {
	nfs := cobrautil.NewNamedFlagSets(cmd)

	dbFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("database"))
	registerDatabaseFlags(dbFlagSet, config.Database)

	authzFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("spicedb"))
	authzFlagSet.StringVar(&config.Authorization.SpiceDBURL, "spicedb-url", config.Authorization.SpiceDBURL, "SpiceDB url")
	authzFlagSet.StringVar(&config.Authorization.PresharedKey, "spicedb-preshared-key", config.Authorization.PresharedKey, "SpiceDB preshared key")

	nfs.AddFlagSets(cmd)
}
