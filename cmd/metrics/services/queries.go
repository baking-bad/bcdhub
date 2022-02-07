package services

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/go-pg/pg/v10"
)

func getScripts(db pg.DBI, lastID int64, size int) (resp []contract.Script, err error) {
	query := db.Model((*contract.Script)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(size).Select(&resp)
	return
}

func getContracts(db pg.DBI, lastID int64, size int) (resp []contract.Contract, err error) {
	var ids []int64
	query := db.Model((*contract.Contract)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("contract.id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	if err = query.Limit(size).Select(&ids); err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	err = db.Model((*contract.Contract)(nil)).WhereIn("contract.id IN (?)", ids).
		Relation("Account").Relation("Manager").Relation("Delegate").Relation("Alpha").Relation("Babylon").
		Select(&resp)
	return
}

func getOperations(db pg.DBI, lastID int64, size int) (resp []operation.Operation, err error) {
	var ids []int64
	query := db.Model((*operation.Operation)(nil)).Column("id")
	if lastID > 0 {
		query.Where("operation.id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	if err = query.Order("operation.id asc").Limit(size).Select(&ids); err != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	err = db.Model((*operation.Operation)(nil)).WhereIn("operation.id IN (?)", ids).
		Relation("Delegate").Relation("Source").Relation("Destination").Relation("Initiator").
		Select(&resp)
	return
}

func getDiffs(db pg.DBI, lastID int64, size int) (resp []bigmapdiff.BigMapDiff, err error) {
	query := db.Model((*bigmapdiff.BigMapDiff)(nil)).Order("id asc")
	if lastID > 0 {
		query.Where("id > ?", lastID)
	}
	if size == 0 || size > 1000 {
		size = 10
	}
	err = query.Limit(size).Select(&resp)
	return
}

func saveSearchModels(ctx context.Context, internalContext *config.Context, items []models.Model) error {
	data := search.Prepare(items)

	return internalContext.Searcher.Save(ctx, data)
}
