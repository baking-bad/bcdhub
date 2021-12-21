package migrations

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/go-pg/pg/v10"
)

// NullableProjectID -
type NullableProjectID struct {
	limit int
}

// NewNullableFields -
func NewNullableProjectID(limit int) *NullableProjectID {
	return &NullableProjectID{limit}
}

// Key -
func (m *NullableProjectID) Key() string {
	return "nullable_project_id"
}

// Description -
func (m *NullableProjectID) Description() string {
	return "set `nullable_project_id` field of `contract` model to nullable type"
}

// Do - migrate function
func (m *NullableProjectID) Do(ctx *config.Context) error {
	if m.limit == 0 {
		m.limit = 10000
	}
	return ctx.StorageDB.DB.RunInTransaction(context.Background(), m.migrateContracts)
}

func (m *NullableProjectID) migrateContracts(tx *pg.Tx) error {
	logger.Info().Msg("processing contracts...")

	var end bool
	var offset int
	for !end {
		var contracts []contract.Contract
		if err := tx.Model(&contract.Contract{}).
			Where("project_id = ''").
			Limit(m.limit).
			Offset(offset).
			Select(&contracts); err != nil {
			return err
		}

		for i := range contracts {
			contracts[i].ProjectID.Valid = false
			if _, err := tx.Model(&contracts[i]).WherePK().Update(); err != nil {
				return err
			}
		}

		offset += len(contracts)
		end = len(contracts) < m.limit
		fmt.Printf("processed %d", offset)
	}

	return nil
}
