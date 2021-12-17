package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/go-pg/pg/v10"
)

func getContracts(db pg.DBI, lastID, size int64) (resp []contract.Contract, err error) {
	query := db.Model((*contract.Contract)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(int(size)).Select(&resp)
	return
}

func getOperations(db pg.DBI, lastID, size int64) (resp []operation.Operation, err error) {
	query := db.Model((*operation.Operation)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(int(size)).Select(&resp)
	return
}

func getDiffs(db pg.DBI, lastID, size int64) (resp []bigmapdiff.BigMapDiff, err error) {
	query := db.Model((*bigmapdiff.BigMapDiff)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(int(size)).Select(&resp)
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
