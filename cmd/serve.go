package cmd

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	v1grpc "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/grpc"
	v1http "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	grpchandlers "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1/grpc"
	v1handlers "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/server"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"github.com/ecordell/optgen/helpers"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/jzelinskie/cobrautil/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ApiVersion string

const (
	ApiV1 ApiVersion = "/api/v1"
)

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
			dt, err := pg.NewPostgresDatastore(ctx, config.Database.URI)
			if err != nil {
				return err
			}
			defer dt.Close()

			// Create v1 server implementation
			v1Server := v1handlers.NewServerV1(dt, config.DataRootFolder)

			var wg sync.WaitGroup
			errCh := make(chan error, 2)

			// Start HTTP server
			wg.Add(1)
			go func() {
				defer wg.Done()
				zap.S().Infof("Starting HTTP server on port %d", config.ServerPort)
				if err := startHTTPServer(ctx, config, v1Server); err != nil {
					errCh <- fmt.Errorf("HTTP server error: %w", err)
				}
			}()

			// Start gRPC server
			wg.Add(1)
			go func() {
				defer wg.Done()
				zap.S().Infof("Starting gRPC server on port %d", config.GrpcPort)
				if err := startGRPCServer(ctx, config, dt); err != nil {
					errCh <- fmt.Errorf("gRPC server error: %w", err)
				}
			}()

			// Wait for either an error or completion
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

// NewServeHTTPCommand creates a command for serving only HTTP
func NewServeHTTPCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve-http",
		Short:        "Serve only the HTTP server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			if err := validateConfig(config); err != nil {
				return err
			}

			// init datastore
			dt, err := pg.NewPostgresDatastore(ctx, config.Database.URI)
			if err != nil {
				return err
			}
			defer dt.Close()

			// Create v1 server implementation
			v1Server := v1handlers.NewServerV1(dt, config.DataRootFolder)

			zap.S().Infof("Starting HTTP server on port %d", config.ServerPort)
			return startHTTPServer(ctx, config, v1Server)
		},
	}

	registerFlags(cmd, config)
	return cmd
}

// NewServeGRPCCommand creates a command for serving only gRPC
func NewServeGRPCCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "serve-grpc",
		Short:        "Serve only the gRPC server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			logger := logger.SetupLogger(config)
			defer logger.Sync()

			undo := zap.ReplaceGlobals(logger)
			defer undo()

			if err := validateConfig(config); err != nil {
				return err
			}

			// init datastore
			dt, err := pg.NewPostgresDatastore(ctx, config.Database.URI)
			if err != nil {
				return err
			}
			defer dt.Close()

			zap.S().Infof("Starting gRPC server on port %d", config.GrpcPort)
			return startGRPCServer(ctx, config, dt)
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
	return nil
}

func startHTTPServer(ctx context.Context, config *config.Config, v1Server *v1handlers.ServerImpl) error {
	httpServer := server.NewRunnableServer(
		server.NewRunnableServerConfigWithOptionsAndDefaults(
			server.WithGraceTimeout(1*time.Second),
			server.WithPort(config.ServerPort),
			server.WithRegisterHandlers(string(ApiV1), func(r *gin.RouterGroup) {
				v1http.RegisterHandlers(r, v1Server)
			}),
			server.WithGinMode(config.GinMode),
			server.WithCloseCallback(func() error {
				zap.S().Info("HTTP server closing")
				return nil
			}),
			server.WithMode(config.Mode),
			server.WithStaticsFolder(config.StaticsFolder),
		),
	)

	httpServer.Run(ctx)
	return nil
}

func startGRPCServer(ctx context.Context, config *config.Config, dt *pg.Datastore) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %d: %w", config.GrpcPort, err)
	}

	// Create filesystem datastore
	fsDatastore := fs.NewFsDatastore(config.DataRootFolder)

	// Create gRPC server with optional logging middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpchandlers.LoggingInterceptor()),
		grpc.StreamInterceptor(grpchandlers.StreamLoggingInterceptor()),
	)

	// Create gRPC handler
	grpcHandler := grpchandlers.NewGRPCServer(dt, fsDatastore)
	v1grpc.RegisterPhotosNGServiceServer(grpcServer, grpcHandler)

	// Enable gRPC reflection for development
	if config.Mode == "dev" {
		reflection.Register(grpcServer)
		zap.S().Info("gRPC reflection enabled for development")
	}

	go func() {
		<-ctx.Done()
		zap.S().Info("gRPC server shutting down...")
		grpcServer.GracefulStop()
	}()

	zap.S().Infof("gRPC server listening on :%d", config.GrpcPort)
	if err := grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
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
	flagSet.IntVar(&config.ServerPort, "http-port", config.ServerPort, "port on which the HTTP server is listening")
	flagSet.IntVar(&config.GrpcPort, "grpc-port", config.GrpcPort, "port on which the gRPC server is listening")
	flagSet.StringVar(&config.GinMode, "server-gin-mode", config.GinMode, "gin mode: either release or debug. It applies only on server-type web")
	flagSet.StringVar(&config.Mode, "server-mode", config.Mode, "server mod: dev or prod")
	flagSet.StringVar(&config.StaticsFolder, "statics-folder", config.StaticsFolder, "path to statics")
	flagSet.StringVar(&config.DataRootFolder, "data-root-folder", config.DataRootFolder, "path to the root folder container media")
}
