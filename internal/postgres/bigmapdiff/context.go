package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
)

func (storage *Storage) buildGetContext(ctx bigmapdiff.GetContext) *gorm.DB {
	query := storage.DB.Table(models.DocBigMapDiff).Select("max(id) as id, count(id) as keys_count")

	if ctx.Network != "" {
		query.Where("network = ?", ctx.Network)
	}
	if ctx.Contract != "" {
		query.Where("contract = ?", ctx.Contract)
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

	size := core.GetPageSize(ctx.Size)
	query.Limit(int(size))

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	return query.Group("key_hash").Order("id desc")
}
