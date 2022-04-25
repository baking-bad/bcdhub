package migrations

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
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
	if err := ctx.Searcher.CreateIndexes(); err != nil {
		return err
	}

	for m.lastID == 0 {
		fmt.Printf("last id = %d\r", m.lastID)
		operations, err := m.getOperations(ctx.StorageDB.DB)
		if err != nil {
			return err
		}
		if err = search.Save(context.Background(), ctx.Searcher, ctx.Network, operations); err != nil {
			return err
		}
		if len(operations) != 100 {
			break
		}
	}
	return nil
}

func (m *FixEntrypointSearch) getOperations(db *pg.DB) (resp []*operation.Operation, err error) {
	query := db.Model((*operation.Operation)(nil)).Order("operation.id asc").
		Relation("Destination").Relation("Source").Relation("Initiator").Relation("Delegate")
	if m.lastID > 0 {
		query.Where("operation.id > ?", m.lastID)
	}
	err = query.Limit(100).Select(&resp)
	return
}
