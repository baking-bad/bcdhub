package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

const (
	statusColumn = "status"
)

// OperationStatusType -
type OperationStatusType struct{}

// Key -
func (m *OperationStatusType) Key() string {
	return "operation_status_type"
}

// Description -
func (m *OperationStatusType) Description() string {
	return "change operation status type string -> int"
}

const (
	viewSeries = `
		create materialized view if not exists series_operation_by_month_%s AS
		with f as (
				select generate_series(
				date_trunc('month', date '2018-06-25'),
				date_trunc('month', now()),
				'1 month'::interval
				) as val
		)
		select
				extract(epoch from f.val),
				count(*) as value
		from f
		left join operations on date_trunc('month', operations.timestamp) = f.val where ((network = %d) and (entrypoint is not null and entrypoint != '') and (status = 1))
		group by 1
		order by date_part
	`
)

// Do - migrate function
func (m *OperationStatusType) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		for i := range ctx.Config.Scripts.Networks {
			dropView := fmt.Sprintf("DROP MATERIALIZED VIEW IF EXISTS series_operation_by_month_%s", ctx.Config.Scripts.Networks[i])
			if err := tx.Exec(dropView).Error; err != nil {
				return err
			}
		}

		if err := m.migrate(tx, new(transfer.Transfer)); err != nil {
			return err
		}
		if err := m.migrate(tx, new(operation.Operation)); err != nil {
			return err
		}

		for i := range ctx.Config.Scripts.Networks {
			script := fmt.Sprintf(viewSeries, ctx.Config.Scripts.Networks[i], types.NewNetwork(ctx.Config.Scripts.Networks[i]))
			if err := tx.Exec(script).Error; err != nil {
				return err
			}
		}
		return nil
	})

}

func (m *OperationStatusType) migrate(tx *gorm.DB, model models.Model) error {
	migrator := tx.Migrator()

	logger.Info("Migrating %s....", model.GetIndex())
	columnTypes, err := migrator.ColumnTypes(model)
	if err != nil {
		return err
	}

	var success bool
	for i := range columnTypes {
		if columnTypes[i].Name() != statusColumn {
			continue
		}
		if columnTypes[i].DatabaseTypeName() != "text" {
			continue
		}

		if err := migrator.RenameColumn(model, statusColumn, bufColumn); err != nil {
			return err
		}
		if err := migrator.AddColumn(model, statusColumn); err != nil {
			return err
		}

		for _, status := range []string{
			"applied", "failed", "skipped", "backtracked",
		} {
			if err := tx.Table(model.GetIndex()).Where("buf_column = ?", status).Update(statusColumn, types.NewOperationStatus(status)).Error; err != nil {
				return err
			}
		}

		if err := migrator.DropColumn(model, bufColumn); err != nil {
			return err
		}

		break
	}

	if success {
		return migrator.AutoMigrate(model)
	}
	return nil
}
