package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"gorm.io/gorm"
)

// EnumToSmallInt - change int64 to small int in database
type EnumToSmallInt struct{}

// Key -
func (m *EnumToSmallInt) Key() string {
	return "enum_to_small_int"
}

// Description -
func (m *EnumToSmallInt) Description() string {
	return "change int64 to small int in database"
}

// Do - migrate function
func (m *EnumToSmallInt) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()

		logger.Info("drop materialized view: head_stats")
		if err := tx.Exec("DROP MATERIALIZED VIEW IF EXISTS head_stats;").Error; err != nil {
			return err
		}

		for _, network := range ctx.Config.Scripts.Networks {
			for _, view := range []string{
				"series_consumed_gas_by_month_%s", "series_contract_by_month_%s", "series_operation_by_month_%s", "series_paid_storage_size_diff_by_month_%s",
			} {
				name := fmt.Sprintf(view, network)
				logger.Info("drop materialized view: %s", name)
				if err := tx.Exec(fmt.Sprintf("DROP MATERIALIZED VIEW IF EXISTS %s;", name)).Error; err != nil {
					return err
				}
			}
		}

		if err := m.alterColumn(migrator, &bigmapaction.BigMapAction{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &bigmapdiff.BigMapState{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &bigmapdiff.BigMapDiff{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &block.Block{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &contract.Contract{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &migration.Migration{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &operation.Operation{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &operation.Operation{}, "status"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &protocol.Protocol{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &tezosdomain.TezosDomain{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &tokenbalance.TokenBalance{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &transfer.Transfer{}, "network"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &transfer.Transfer{}, "status"); err != nil {
			return err
		}
		if err := m.alterColumn(migrator, &tzip.TZIP{}, "network"); err != nil {
			return err
		}

		return nil
	})
}

func (m *EnumToSmallInt) alterColumn(migrator gorm.Migrator, model interface{}, column string) error {
	if data, ok := model.(models.Model); ok {
		logger.Info("Migrating column '%s' of '%s'", column, data.GetIndex())
	}

	if !migrator.HasColumn(model, column) {
		return nil
	}

	columnTypes, err := migrator.ColumnTypes(model)
	if err != nil {
		return err
	}

	for i := range columnTypes {
		if columnTypes[i].Name() != column {
			continue
		}

		if columnTypes[i].DatabaseTypeName() == "int2" {
			break
		}

		if err := migrator.AlterColumn(model, "network"); err != nil {
			return err
		}
		break
	}

	return nil
}
