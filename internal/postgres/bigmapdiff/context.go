package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/go-pg/pg/v10/orm"
)

func (storage *Storage) buildGetContext(ctx bigmapdiff.GetContext) *orm.Query {
	query := storage.DB.Model().Table(models.DocBigMapDiff).ColumnExpr("max(id) as id, count(id) as keys_count")

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

	query.Limit(storage.GetPageSize(ctx.Size))

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	return query.Group("key_hash").Order("id desc")
}

func (storage *Storage) buildGetContextForState(ctx bigmapdiff.GetContext) *orm.Query {
	query := storage.DB.Model().Table(models.DocBigMapState)

	if ctx.Contract != "" {
		query.Where("contract = ?", ctx.Contract)
	}
	if ctx.Ptr != nil {
		query.Where("ptr = ?", *ctx.Ptr)
	}
	if ctx.MaxLevel != nil {
		query.Where("last_update_level < ?", *ctx.MaxLevel)
	}
	if ctx.MinLevel != nil {
		query.Where("last_update_level >= ?", *ctx.MinLevel)
	}
	if ctx.CurrentLevel != nil {
		query.Where("last_update_level = ?", *ctx.CurrentLevel)
	}

	query.Limit(storage.GetPageSize(ctx.Size))

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	return query.Order("id desc")
}
