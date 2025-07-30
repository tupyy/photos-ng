package pg

// copyright SpiceDB

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryFilter func(original sq.SelectBuilder) sq.SelectBuilder

type TxUserFunc func(context.Context, *Writer) error

type Datastore struct {
	pool ConnPooler
}

// NewPostgresDatastore creates a new Postgres datastore instance with the given configuration options.
// It establishes a connection pool and sets up query interceptors for logging and monitoring.
func NewPostgresDatastore(ctx context.Context, url string, options ...Option) (*Datastore, error) {
	pgOptions := newPostgresConfig(options)

	pgxConfig, err := pgOptions.PgxConfig(url)
	if err != nil {
		return nil, err
	}

	pgPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	if err := pgPool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Datastore{pool: MustNewInterceptorPooler(pgPool, newLogInterceptor())}, nil
}

func (dt *Datastore) Close() {
	dt.pool.Close()
}
