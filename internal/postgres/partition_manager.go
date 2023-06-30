package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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
func NewPartitionManager(conn *core.Postgres) PartitionManager {
	return PartitionManager{
		conn: conn,
	}
}

const createPartitionTemplate = `CREATE TABLE IF NOT EXISTS ? PARTITION OF ? FOR VALUES FROM (?) TO (?);`

func quarterOf(month time.Month) int {
	return (int(month) + 2) / 3
}

func quarterBoundaries(current time.Time) (time.Time, time.Time, error) {
	year := current.Year()
	quarter := quarterOf(current.Month())

	switch quarter {
	case 1:
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 2:
		start := time.Date(year, time.April, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 3:
		start := time.Date(year, time.July, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	case 4:
		start := time.Date(year, time.October, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0)
		return start, end, nil
	}

	return time.Now(), time.Now(), errors.New("invalid quarter")
}

func (pm *PartitionManager) partitionId(currentTime time.Time) string {
	return fmt.Sprintf("%dQ%d", currentTime.Year(), quarterOf(currentTime.Month()))
}

// CreatePartitions -
func (pm *PartitionManager) CreatePartitions(ctx context.Context, currentTime time.Time) error {
	id := pm.partitionId(currentTime)
	if id == pm.lastId {
		return nil
	}

	start, end, err := quarterBoundaries(currentTime)
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
