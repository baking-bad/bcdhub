package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
	"gorm.io/gorm"
)

func getContracts(db *gorm.DB, table string, lastID, size int64) (resp []contract.Contract, err error) {
	query := db.Table(table).Order("id asc").Where("project_id = ''")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(int(size)).Find(&resp).Error
	return
}

func getOperations(db *gorm.DB, table string, lastID, size int64) (resp []operation.Operation, err error) {
	query := db.Table(table).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(int(size)).Find(&resp).Error
	return
}

func saveSearchModels(ctx *config.Context, items []models.Model) error {
	data := search.Prepare(items)

	for i := range data {
		switch typ := data[i].(type) {
		case *search.Contract:
			typ.Alias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Address)
			typ.DelegateAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Delegate)
		case *search.Operation:
			typ.SourceAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Source)
			typ.DestinationAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Destination)
			typ.DelegateAlias = ctx.CachedAlias(types.NewNetwork(typ.Network), typ.Delegate)
		}
	}

	return ctx.Searcher.Save(data)
}
