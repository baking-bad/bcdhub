package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
)

func buildGetContext(query *gorm.DB, ctx bigmapdiff.GetContext, withGroup bool) {
	if query == nil {
		return
	}

	if ctx.Network != "" {
		query.Where("network = ?", ctx.Network)
	}
	if ctx.Contract != "" {
		query.Where("address = ?", ctx.Contract)
	}
	if ctx.Ptr != nil {
		query.Where("ptr = ?", *ctx.Ptr)
	}
	if ctx.MaxLevel != nil {
		query.Where("level < ?", *ctx.MaxLevel)
	}
	if ctx.MinLevel != nil {
		query.Where("level >= ?", *ctx.MinLevel)
	}
	if ctx.CurrentLevel != nil {
		query.Where("level = ?", *ctx.CurrentLevel)
	}
	if ctx.Query != "" {
		query.Where("key_hash LIKE %?%", ctx.Query)
	}

	if ctx.Size > 0 {
		size := core.GetPageSize(ctx.Size)
		query.Limit(int(size))
	}

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	if withGroup {
		query.Group("key_hash")
	}
	query.Order("indexed_time desc")
}
