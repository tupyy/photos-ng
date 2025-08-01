package cmd

import (
	"context"
	"fmt"
	"time"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	v1handlers "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"github.com/ecordell/optgen/helpers"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/jzelinskie/cobrautil/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type ApiVersion string

const (
	ApiV1 ApiVersion = "/api/v1"
)

// NewServeCommand creates a new cobra command for starting the server.
// It sets up the database connection, configures the HTTP server, and handles graceful shutdown.
func NewServeCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve",
		Short:        "Serve the server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			zap.S().Info("using configuration", "config", helpers.Flatten(config.DebugMap()))

			if config.Mode == "prod" && config.StaticsFolder == "" {
				return fmt.Errorf("statics folder should be provided in prod mod")
			}

			// init datastore
			dt, err := pg.NewPostgresDatastore(context.Background(), config.Database.URI)
			if err != nil {
				return err
			}

			// Create v1 server implementation
			v1Server := v1handlers.NewServerV1(dt, config.DataRootFolder)

			server := server.NewRunnableServer(
				server.NewRunnableServerConfigWithOptionsAndDefaults(
					server.WithGraceTimeout(1*time.Second),
					server.WithPort(config.ServerPort),
					server.WithRegisterHandlers(string(ApiV1), func(r *gin.RouterGroup) {
						// Register the generated v1 API handlers
						v1.RegisterHandlers(r, v1Server)
					}),
					server.WithGinMode(config.GinMode),
					server.WithCloseCallback(func() error {
						zap.S().Info("close datastore")
						dt.Close()
						return nil
					}),
					server.WithMode(config.Mode),
					server.WithStaticsFolder(config.StaticsFolder),
				),
			)

			server.Run(context.Background())

			return nil
		},
	}

	registerFlags(cmd, config)

	return cmd
}

func registerFlags(cmd *cobra.Command, config *config.Config) {
	nfs := cobrautil.NewNamedFlagSets(cmd)

	dbFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("database"))
	registerDatabaseFlags(dbFlagSet, config.Database)

	serverFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("server"))
	registerServerFlags(serverFlagSet, config)

	nfs.AddFlagSets(cmd)
}

func registerDatabaseFlags(flagSet *pflag.FlagSet, config *config.Database) {
	flagSet.StringVar(&config.URI, "db-conn-uri", config.URI, `connection string used by remote databases (e.g. "postgres://postgres:password@localhost:5432/photos")`)
	flagSet.BoolVar(&config.SSL, "db-ssl-mode", config.SSL, "ssl mode")
}

func registerServerFlags(flagSet *pflag.FlagSet, config *config.Config) {
	flagSet.IntVar(&config.ServerPort, "server-port", config.ServerPort, "port on which the server is listening")
	flagSet.StringVar(&config.GinMode, "server-gin-mode", config.GinMode, "gin mode: either release or debug. It applies only on server-type web")
	flagSet.StringVar(&config.Mode, "server-mode", config.Mode, "server mod: dev or prod")
	flagSet.StringVar(&config.StaticsFolder, "statics-folder", config.StaticsFolder, "path to statics")
	flagSet.StringVar(&config.DataRootFolder, "data-root-folder", config.DataRootFolder, "path to the root folder container media")
}
