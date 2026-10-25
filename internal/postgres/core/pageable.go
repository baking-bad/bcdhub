package core

import "github.com/baking-bad/bcdhub/internal/postgres/consts"

// GetPageSize - validate and return page size
func (pg *Postgres) GetPageSize(size int64) int64 {
	switch {
	case size > consts.MaxSize:
		return consts.MaxSize
	case size == 0:
		return pg.PageSize
	default:
		return size
	}
}
