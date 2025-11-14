package cmd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ecordell/optgen/helpers"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/jzelinskie/cobrautil/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	httpv1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	v1grpc "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1/grpc"
	v1http "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/datastore"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/spicedb"
)

type ApiVersion string

const (
	ApiV1 ApiVersion = "/api/v1"
)

type Server interface {
	Start(context.Context) error
	Stop(context.Context, chan any)
}

// NewServeCommand creates a new cobra command for starting both HTTP and gRPC servers.
func NewServeCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve",
		Short:        "Serve both HTTP and gRPC servers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			zap.S().Info("using configuration", "config", helpers.Flatten(config.DebugMap()))

			if err := validateConfig(config); err != nil {
				return err
			}

			// init datastore
			pgDatastore, err := datastore.NewConnPool(ctx, config.Database.URI)
			if err != nil {
				return err
			}
			defer pgDatastore.Close()

			// Create v1 handlers for http and grpc
			httpHandler := v1http.NewHandler(pg.NewPostgresDatastore(pgDatastore), fs.NewFsDatastore(config.DataRootFolder))
			grpcHandler := v1grpc.NewHandler(pg.NewPostgresDatastore(pgDatastore), fs.NewFsDatastore(config.DataRootFolder))

			if config.Authorization.Enabled {
				spiceClient, err := spicedb.InitSpiceDBClient(config.Authorization.SpiceDBURL, config.Authorization.PresharedKey)
				if err != nil {
					return err
				}
				httpHandler = v1http.NewHandlerWithAuthorization(spiceClient, pg.NewPostgresDatastore(pgDatastore), fs.NewFsDatastore(config.DataRootFolder))
				grpcHandler = v1grpc.NewHandlerWithAuthorization(spiceClient, pg.NewPostgresDatastore(pgDatastore), fs.NewFsDatastore(config.DataRootFolder))
			}

			var wg sync.WaitGroup
			errCh := make(chan error, 2)

			// Start HTTP server
			wg.Add(1)
			go func() {
				defer wg.Done()
				zap.S().Infof("Starting HTTP server on port %d", config.HttpPort)
				cfg := server.NewHttpServerConfigWithOptionsAndDefaults(
					server.WithGraceTimeout(1*time.Second),
					server.WithHttpPort(config.HttpPort),
					server.WithRegisterHandlers(string(ApiV1), func(r *gin.RouterGroup) {
						httpv1.RegisterHandlers(r, httpHandler)
					}),
					server.WithGinMode(config.GinMode),
					server.WithMode(config.Mode),
					server.WithStaticsFolder(config.StaticsFolder),
				)

				if config.Authentication.Enabled {
					cfg = cfg.WithOptions(server.WithAuthentication(server.Authentication{
						WellknownURL: config.Authentication.WellknownURL,
					}))
				}

				srv, err := server.NewHttpServer(cfg)
				if err != nil {
					errCh <- fmt.Errorf("HTTP server error: %w", err)
					return
				}
				if err := srv.Start(ctx); err != nil {
					errCh <- fmt.Errorf("HTTP server error: %w", err)
				}
			}()

			// Start gRPC server
			wg.Add(1)
			go func() {
				defer wg.Done()
				cfg := server.NewGrpcServerConfigWithOptionsAndDefaults(
					server.WithGrpcHandler(grpcHandler),
					server.WithGrpcPort(config.GrpcPort),
					server.WithGrpcMode("dev"),
				)
				srv, err := server.NewGrpcServer(cfg)
				if err != nil {
					errCh <- fmt.Errorf("GRPC server error: %w", err)
				}
				if err := srv.Start(ctx); err != nil {
					errCh <- fmt.Errorf("GRPC server error: %w", err)
				}
				zap.S().Infof("Starting gRPC server on port %d", config.GrpcPort)
			}()

			go func() {
				wg.Wait()
				close(errCh)
			}()

			select {
			case err := <-errCh:
				if err != nil {
					cancel()
					return err
				}
			case <-ctx.Done():
				zap.S().Info("Servers shutting down...")
			}

			return nil
		},
	}

	registerFlags(cmd, config)
	return cmd
}

func validateConfig(config *config.Config) error {
	if config.Mode == "prod" && config.StaticsFolder == "" {
		return fmt.Errorf("statics folder should be provided in prod mode")
	}
	if config.DataRootFolder == "" {
		return fmt.Errorf("data root folder cannot be empty")
	}

	if config.Authentication.Enabled {
		if config.Authentication.WellknownURL == "" {
			return errors.New("wellknown_url cannot be empty")
		}
	}

	if config.Authorization.Enabled {
		if config.Authorization.SpiceDBURL == "" {
			return errors.New("spicedb url cannot be empty")
		}
		if config.Authorization.PresharedKey == "" {
			return errors.New("preshared key cannot be empty")
		}
	}
	return nil
}

func registerFlags(cmd *cobra.Command, config *config.Config) {
	nfs := cobrautil.NewNamedFlagSets(cmd)

	dbFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("database"))
	registerDatabaseFlags(dbFlagSet, config.Database)

	serverFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("server"))
	registerServerFlags(serverFlagSet, config)

	authenticationFlagSet := nfs.FlagSet(color.New(color.FgBlue, color.Bold).Sprint("authentication"))
	registerAuthenticationFlags(authenticationFlagSet, config)

	authorizationFlatSet := nfs.FlagSet(color.New(color.FgBlue, color.Bold).Sprint("authorization"))
	registerAuthorizationFlags(authorizationFlatSet, config)

	nfs.AddFlagSets(cmd)
}

func registerDatabaseFlags(flagSet *pflag.FlagSet, config *config.Database) {
	flagSet.StringVar(&config.URI, "db-conn-uri", config.URI, `connection string used by remote databases (e.g. "postgres://postgres:password@localhost:5432/photos")`)
	flagSet.BoolVar(&config.SSL, "db-ssl-mode", config.SSL, "ssl mode")
}

func registerServerFlags(flagSet *pflag.FlagSet, config *config.Config) {
	flagSet.IntVar(&config.HttpPort, "http-port", config.HttpPort, "port on which the HTTP server is listening")
	flagSet.IntVar(&config.GrpcPort, "grpc-port", config.GrpcPort, "port on which the gRPC server is listening")
	flagSet.StringVar(&config.GinMode, "server-gin-mode", config.GinMode, "gin mode: either release or debug. It applies only on server-type web")
	flagSet.StringVar(&config.Mode, "server-mode", config.Mode, "server mod: dev or prod")
	flagSet.StringVar(&config.StaticsFolder, "statics-folder", config.StaticsFolder, "path to statics")
	flagSet.StringVar(&config.DataRootFolder, "data-root-folder", config.DataRootFolder, "path to the root folder container media")
}

func registerAuthenticationFlags(flagSet *pflag.FlagSet, config *config.Config) {
	flagSet.BoolVar(&config.Authentication.Enabled, "authentication-enabled", config.Authentication.Enabled, "enable OIDC authentication (default false)")
	flagSet.StringVar(&config.Authentication.WellknownURL, "authentication-wellknown-endpoint", config.Authentication.WellknownURL, "OIDC provider wellknown endpoing address")
}

func registerAuthorizationFlags(flagSet *pflag.FlagSet, config *config.Config) {
	flagSet.BoolVar(&config.Authorization.Enabled, "authorization-enabled", config.Authorization.Enabled, "enable authorization")
	flagSet.StringVar(&config.Authorization.SpiceDBURL, "authorization-spicedb-url", config.Authorization.SpiceDBURL, "SpiceDB url")
	flagSet.StringVar(&config.Authorization.PresharedKey, "authorization-spicedb-preshared-key", config.Authorization.PresharedKey, "SpiceDB preshared key")
}
