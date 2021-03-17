package core

import "github.com/baking-bad/bcdhub/internal/postgres/consts"

// GetPageSize - validate and return page size
func GetPageSize(size int64) int {
	switch {
	case size > consts.MaxSize:
		return consts.MaxSize
	case size == 0:
		return consts.DefaultSize
	default:
		return int(size)
	}
}
