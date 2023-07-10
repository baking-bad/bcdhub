package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
)

// PartitionManager -
type PartitionManager struct {
	conn *core.Postgres

	lastId string
}

// NewPartitionManager -
func NewPartitionManager(conn *core.Postgres) *PartitionManager {
	return &PartitionManager{
		conn: conn,
	}
}

const createPartitionTemplate = `CREATE TABLE IF NOT EXISTS ? PARTITION OF ? FOR VALUES FROM (?) TO (?);`

func (pm *PartitionManager) partitionId(currentTime time.Time) string {
	return fmt.Sprintf("%dQ%d", currentTime.Year(), helpers.QuarterOf(currentTime.Month()))
}

// CreatePartitions -
func (pm *PartitionManager) CreatePartitions(ctx context.Context, currentTime time.Time) error {
	id := pm.partitionId(currentTime)
	if id == pm.lastId {
		return nil
	}

	start, end, err := helpers.QuarterBoundaries(currentTime)
	if err != nil {
		return err
	}

	for _, model := range []models.Model{
		&operation.Operation{},
		&bigmapdiff.BigMapDiff{},
	} {
		partitionName := fmt.Sprintf("%s_%s", model.GetIndex(), id)
		if _, err := pm.conn.DB.ExecContext(
			ctx,
			createPartitionTemplate,
			pg.Ident(partitionName),
			pg.Ident(model.GetIndex()),
			start.Format(time.RFC3339Nano),
			end.Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}

	pm.lastId = id
	return nil
}
