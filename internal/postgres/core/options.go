package core

import (
	"time"

	bcdLogger "github.com/baking-bad/bcdhub/internal/logger"
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
		sql, err := pg.DB.DB()
		if err != nil {
			bcdLogger.Err(err)
			return
		}
		sql.SetMaxOpenConns(count)
		sql.SetConnMaxLifetime(time.Hour)
	}
}

// WithIdleConnections -
func WithIdleConnections(count int) PostgresOption {
	return func(pg *Postgres) {
		if count == 0 {
			count = consts.DefaultSize
		}
		sql, err := pg.DB.DB()
		if err != nil {
			bcdLogger.Err(err)
			return
		}
		sql.SetMaxIdleConns(count)
		sql.SetConnMaxIdleTime(time.Minute * 30)
	}
}
