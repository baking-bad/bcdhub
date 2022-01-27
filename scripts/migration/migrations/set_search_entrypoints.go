package migrations

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/go-pg/pg/v10"
)

// FixEntrypointSearch -
type FixEntrypointSearch struct {
	lastID int64
}

// Key -
func (m *FixEntrypointSearch) Key() string {
	return "fix_entrypoint_search"
}

// Description -
func (m *FixEntrypointSearch) Description() string {
	return "fill `operations` index in elasticsearch"
}

// Do - migrate function
func (m *FixEntrypointSearch) Do(ctx *config.Context) error {
	var err error
	operations := make([]operation.Operation, 0)

	if err := ctx.Searcher.CreateIndexes(); err != nil {
		return err
	}

	for m.lastID == 0 || len(operations) == 1000 {
		fmt.Printf("last id = %d\r", m.lastID)
		operations, err = m.getOperations(ctx.StorageDB.DB)
		if err != nil {
			return err
		}
		if err = m.saveSearchModels(ctx, operations); err != nil {
			return err
		}
	}
	return nil
}

func (m *FixEntrypointSearch) getOperations(db *pg.DB) (resp []operation.Operation, err error) {
	query := db.Model((*operation.Operation)(nil)).Order("operation.id asc").
		Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate")
	if m.lastID > 0 {
		query.Where("operation.id > ?", m.lastID)
	}
	err = query.Limit(1000).Select(&resp)
	return
}

func (m *FixEntrypointSearch) saveSearchModels(internalContext *config.Context, operations []operation.Operation) error {
	items := make([]models.Model, len(operations))
	for i := range operations {
		items[i] = &operations[i]
		if m.lastID < operations[i].ID {
			m.lastID = operations[i].ID
		}
	}
	return internalContext.Searcher.Save(context.Background(), search.Prepare(items))
}
