package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Nerzal/gocloak/v13"
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

func NewAuthzMigrateCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "migrate-authz",
		Short:        "Run spicedb migrations",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAuthzMigrateConfig(config); err != nil {
				return err
			}

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

			keycloakBaseURL, realm, err := parseWellknownURL(config.Authentication.WellknownURL)
			if err != nil {
				return err
			}
			client := gocloak.NewClient(keycloakBaseURL)
			// Run migrations
			zap.S().Info("running spicedb migrations")
			if err := migrations.NewAuthzMigrations().Migrate(context.Background(), services.NewUserService(client, config.Authentication.ClientID, config.Authentication.ClientSecret, realm), authServ); err != nil {
				zap.S().Error("migration failed", "error", err)
				return err
			}

			zap.S().Info("spicedb migrations completed successfully")
			return nil
		},
	}

	registerAuthzMigrateFlags(cmd, config)
	return cmd
}

func registerAuthzMigrateFlags(cmd *cobra.Command, config *config.Config) {
	nfs := cobrautil.NewNamedFlagSets(cmd)

	dbFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("database"))
	registerDatabaseFlags(dbFlagSet, config.Database)

	authenticationFlagSet := nfs.FlagSet(color.New(color.FgBlue, color.Bold).Sprint("authentication"))
	registerAuthenticationFlags(authenticationFlagSet, config)

	authzFlagSet := nfs.FlagSet(color.New(color.FgCyan, color.Bold).Sprint("spicedb"))
	authzFlagSet.StringVar(&config.Authorization.SpiceDBURL, "spicedb-url", config.Authorization.SpiceDBURL, "SpiceDB url")
	authzFlagSet.StringVar(&config.Authorization.PresharedKey, "spicedb-preshared-key", config.Authorization.PresharedKey, "SpiceDB preshared key")

	nfs.AddFlagSets(cmd)
}

func validateAuthzMigrateConfig(config *config.Config) error {
	if config.Database.URI == "" {
		return errors.New("--db-conn-uri cannot be empty")
	}
	if config.Authentication.WellknownURL == "" {
		return errors.New("--authentication-wellknown-endpoint cannot be empty")
	}
	if config.Authentication.ClientID == "" {
		return errors.New("--authentication-client-id cannot be empty")
	}
	if config.Authentication.ClientSecret == "" {
		return errors.New("--authentication-client-secret cannot be empty")
	}
	if config.Authorization.SpiceDBURL == "" {
		return errors.New("--spicedb-url cannot be empty")
	}
	if config.Authorization.PresharedKey == "" {
		return errors.New("--spicedb-preshared-key cannot be empty")
	}
	return nil
}

// parseWellknownURL extracts the base URL and realm from an OIDC wellknown URL.
// Expected format: {scheme}://{host}/realms/{realm}/.well-known/openid-configuration
func parseWellknownURL(wellknownURLStr string) (baseURL, realm string, err error) {
	parsed, err := url.Parse(wellknownURLStr)
	if err != nil {
		return "", "", err
	}
	baseURL = parsed.Scheme + "://" + parsed.Host

	pathParts := strings.Split(parsed.Path, "/")
	for i, part := range pathParts {
		if part == "realms" && i+1 < len(pathParts) {
			realm = pathParts[i+1]
			break
		}
	}
	if realm == "" {
		return "", "", fmt.Errorf("realm not found in wellknown URL: %s", wellknownURLStr)
	}
	return baseURL, realm, nil
}
