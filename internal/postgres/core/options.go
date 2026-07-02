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

// WithQueryLogging -
func WithQueryLogging() PostgresOption {
	return func(pg *Postgres) {
		pg.hasLogger = true
	}
}

func WithTimeout(timeout time.Duration) PostgresOption {
	return func(pg *Postgres) {
		pg.timeout = timeout
	}
}
