package migrations

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

// MigrateStore runs database migrations using the goose migration tool.
// It takes a database connection and migration folder path, and applies
// all pending migrations in the specified folder.
func MigrateStore(db *sql.DB, migrationFolder string) error {
	goose.SetLogger(&logger{})

	fi, err := os.Stat(migrationFolder)
	if err != nil {
		return err
	}

	if !fi.Mode().IsDir() {
		return fmt.Errorf("failed to open migration folder: %s is not a folder", migrationFolder)
	}

	goose.SetBaseFS(os.DirFS(migrationFolder))

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}

/*
logger implements goose.Logger interface

	type Logger interface {
		Fatalf(format string, v ...interface{})
		Printf(format string, v ...interface{})
	}
*/
type logger struct{}

func (m *logger) Printf(format string, v ...interface{}) {}
func (m *logger) Fatalf(format string, v ...interface{}) {}
