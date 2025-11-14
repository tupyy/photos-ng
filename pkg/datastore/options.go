package datastore

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolOptions struct {
	ConnMaxIdleTime         time.Duration
	ConnMaxLifetime         time.Duration
	ConnMaxLifetimeJitter   time.Duration
	ConnHealthCheckInterval time.Duration
	MinOpenConns            int
	MaxOpenConns            int
}

func defaultPoolOptions() PoolOptions {
	return PoolOptions{
		ConnMaxIdleTime:         30 * time.Minute,
		ConnMaxLifetime:         30 * time.Minute,
		MaxOpenConns:            10,
		MinOpenConns:            10,
		ConnHealthCheckInterval: 30 * time.Second,
	}
}

type Option func(*postgresOptions)

type postgresOptions struct {
	poolOptions PoolOptions
}

func (c *postgresOptions) PgxConfig(url string) (*pgxpool.Config, error) {
	pgxConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pgxConfig.MaxConnIdleTime = c.poolOptions.ConnMaxIdleTime
	pgxConfig.MaxConnLifetime = c.poolOptions.ConnMaxLifetime
	pgxConfig.MaxConns = int32(c.poolOptions.MaxOpenConns)
	pgxConfig.MinConns = int32(c.poolOptions.MinOpenConns)
	pgxConfig.HealthCheckPeriod = c.poolOptions.ConnHealthCheckInterval

	return pgxConfig, nil
}

func newPostgresConfig(options []Option) *postgresOptions {
	pgOptions := &postgresOptions{
		poolOptions: defaultPoolOptions(),
	}

	for _, option := range options {
		option(pgOptions)
	}

	return pgOptions
}

// ConnMaxIdleTime is the duration after which an idle read connection will
// be automatically closed by the health check.
//
// This value defaults to having no maximum.
func ConnMaxIdleTime(idle time.Duration) Option {
	return func(po *postgresOptions) { po.poolOptions.ConnMaxIdleTime = idle }
}

// ConnMaxLifetime is the duration since creation after which a read
// connection will be automatically closed.
//
// This value defaults to having no maximum.
func ConnMaxLifetime(lifetime time.Duration) Option {
	return func(po *postgresOptions) { po.poolOptions.ConnMaxLifetime = lifetime }
}

func ConnHealthCheckInterval(interval time.Duration) Option {
	return func(po *postgresOptions) { po.poolOptions.ConnHealthCheckInterval = interval }
}

// ConnMaxLifetimeJitter is an interval to wait up to after the max lifetime
// to close the connection.
//
// This value defaults to 20% of the max lifetime.
func ConnMaxLifetimeJitter(jitter time.Duration) Option {
	return func(po *postgresOptions) { po.poolOptions.ConnMaxLifetimeJitter = jitter }
}

// ConnsMinOpen is the minimum size of the connection pool used for reads.
//
// The health check will increase the number of connections to this amount if
// it had dropped below.
//
// This value defaults to the maximum open connections.
func ConnsMinOpen(conns int) Option {
	return func(po *postgresOptions) { po.poolOptions.MinOpenConns = conns }
}

// ConnsMaxOpen is the maximum size of the connection pool used for reads.
//
// This value defaults to having no maximum.
func ConnsMaxOpen(conns int) Option {
	return func(po *postgresOptions) { po.poolOptions.MaxOpenConns = conns }
}
