package core

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/postgres/consts"
)

// PostgresOption -
type PostgresOption func(pg *Postgres)

// WithPageSize -
func WithPageSize(pageSize int64) PostgresOption {
	return func(pg *Postgres) {
		if pageSize == 0 {
			pageSize = consts.DefaultSize
		}
		pg.PageSize = pageSize
	}
}

// WithMaxConnections -
func WithMaxConnections(count int) PostgresOption {
	return func(pg *Postgres) {
		if count == 0 {
			count = consts.DefaultSize
		}
		if opts := pg.DB.Options(); opts != nil {
			opts.PoolSize = count
			opts.MaxConnAge = time.Hour
		}
	}
}

// WithIdleConnections -
func WithIdleConnections(count int) PostgresOption {
	return func(pg *Postgres) {
		if count == 0 {
			count = consts.DefaultSize
		}
		if opts := pg.DB.Options(); opts != nil {
			opts.IdleTimeout = time.Minute * 30
			opts.MinIdleConns = count
		}
	}
}

// WithQueryLogging -
func WithQueryLogging() PostgresOption {
	return func(pg *Postgres) {
		if pg.DB == nil {
			return
		}
		pg.DB.AddQueryHook(&logQueryHook{})
	}
}
