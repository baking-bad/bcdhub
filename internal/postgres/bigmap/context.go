package bigmap

import (
	"database/sql"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"gorm.io/gorm"
)

func buildGetContext(tx *gorm.DB, ctx bigmap.GetContext, size int) *gorm.DB {
	query := tx.Table(models.DocBigMapDiff).
		Joins("left join big_maps on big_maps.id = big_map_id").
		Select("max(big_map_diffs.id) as id, count(big_map_diffs.id) as keys_count")

	if ctx.Network != 0 {
		query.Where("big_maps.network = ?", ctx.Network)
	}
	if ctx.Contract != "" {
		query.Where("big_maps.contract = ?", ctx.Contract)
	}
	if ctx.Ptr != nil {
		query.Where("big_maps.ptr = ?", *ctx.Ptr)
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
		query.Where("(key_hash LIKE @reg OR array_to_string(key_strings, '|') LIKE @reg)", sql.Named("reg", fmt.Sprintf("%%%s%%", ctx.Query)))
	}

	query.Limit(size)

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	return query.Group("key_hash").Order("id desc")
}

func buildGetContextForState(tx *gorm.DB, ctx bigmap.GetContext, size int) *gorm.DB {
	query := tx.Table(models.DocBigMapState).
		Joins("left join big_maps on big_maps.id = big_map_id")

	if ctx.Network != 0 {
		query.Where("big_maps.network = ?", ctx.Network)
	}
	if ctx.Contract != "" {
		query.Where("big_maps.contract = ?", ctx.Contract)
	}
	if ctx.Ptr != nil {
		query.Where("big_maps.ptr = ?", *ctx.Ptr)
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
		query.Where("(key_hash LIKE @reg)", sql.Named("reg", fmt.Sprintf("%%%s%%", ctx.Query)))
	}

	query.Limit(size)

	if ctx.Offset > 0 {
		query.Offset(int(ctx.Offset))
	}

	return query.Order("big_map_states.id desc")
}
