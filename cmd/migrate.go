package cmd

import (
	"database/sql"
	"path/filepath"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg/migrations"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"github.com/fatih/color"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jzelinskie/cobrautil/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewMigrateCommand creates a new cobra command for running database migrations.
func NewMigrateCommand(config *config.Config) *cobra.Command {
	var migrationPath string

	cmd := &cobra.Command{
		Use:          "migrate",
		Short:        "Run database migrations",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			zap.S().Infow("starting database migration", "db_uri", config.Database.URI, "migration_path", migrationPath)

			// Connect to database using stdlib sql interface
			db, err := sql.Open("pgx", config.Database.URI)
			if err != nil {
				zap.S().Error("failed to connect to database", "error", err)
				return err
			}
			defer db.Close()

			// Test connection
			if err := db.Ping(); err != nil {
				zap.S().Error("failed to ping database", "error", err)
				return err
			}

			zap.S().Info("connected to database successfully")

			// Resolve absolute path for migrations
			absPath, err := filepath.Abs(migrationPath)
			if err != nil {
				zap.S().Error("failed to resolve migration path", "error", err)
				return err
			}

			// Run migrations
			zap.S().Info("running migrations", "path", absPath)
			if err := migrations.MigrateStore(db, absPath); err != nil {
				zap.S().Error("migration failed", "error", err)
				return err
			}

			zap.S().Info("migrations completed successfully")
			return nil
		},
	}

	registerMigrateFlags(cmd, config, &migrationPath)
	return cmd
}

func registerMigrateFlags(cmd *cobra.Command, config *config.Config, migrationPath *string) {
	nfs := cobrautil.NewNamedFlagSets(cmd)

	dbFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("database"))
	registerDatabaseFlags(dbFlagSet, config.Database)

	migrationFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("migration"))

	migrationFlagSet.StringVar(migrationPath, "migration-folder", "", "path to database migration files (required)")
	cmd.MarkFlagRequired("migration-folder")

	nfs.AddFlagSets(cmd)
}
