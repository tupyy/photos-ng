package main

import (
	"os"

	"git.tls.tupangiu.ro/cosmin/photos-ng/cmd"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	cfg := config.NewConfigWithOptionsAndDefaults(
		config.WithDatabase(config.NewDatabaseWithOptions(
			config.WithURI("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
			config.WithMaxOpenConnections(10),
		)),
		config.WithServerPort(8080),
		config.WithLogFormat("console"),
		config.WithLogLevel("debug"),
		config.WithGinMode("debug"),
	)

	var rootCmd = &cobra.Command{
		Use:   "finance",
		Short: "Manage my finances",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}
	registerLoggingFlags(rootCmd, cfg)

	rootCmd.AddCommand(cmd.NewServeCommand(cfg))
	rootCmd.AddCommand(cmd.NewMigrateCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func registerLoggingFlags(cmd *cobra.Command, config *config.Config) {
	cmd.PersistentFlags().StringVar(&config.LogFormat, "log-format", config.LogFormat, "format of the logs: console or json")
	cmd.PersistentFlags().StringVar(&config.LogLevel, "log-level", config.LogLevel, "log level")
}
