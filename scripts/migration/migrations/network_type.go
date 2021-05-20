package migrations

import (
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
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"gorm.io/gorm"
)

const (
	bufColumn     = "buf_column"
	networkColumn = "network"
)

var namesToNetwork = map[string]types.Network{
	"mainnet":     types.Mainnet,
	"carthagenet": types.Carthagenet,
	"delphinet":   types.Delphinet,
	"edo2net":     types.Edo2net,
	"florencenet": types.Florencenet,
}

// NetworkType -
type NetworkType struct{}

// Key -
func (m *NetworkType) Key() string {
	return "network_type"
}

// Description -
func (m *NetworkType) Description() string {
	return "change network type string -> int"
}

// Do - migrate function
func (m *NetworkType) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DROP MATERIALIZED VIEW IF EXISTS public.head_stats;`).Error; err != nil {
			return err
		}

		if err := m.migrate(tx, new(bigmapaction.BigMapAction)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(bigmapdiff.BigMapDiff)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(bigmapdiff.BigMapState)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(block.Block)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(contract.Contract)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(migration.Migration)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(operation.Operation)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(protocol.Protocol)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(tezosdomain.TezosDomain)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(tokenbalance.TokenBalance), "token_balances_token_idx"); err != nil {
			return err
		}

		if err := m.migrate(tx, new(tokenmetadata.TokenMetadata)); err != nil {
			return err
		}

		if err := m.migrate(tx, new(transfer.Transfer), "transfers_network_idx"); err != nil {
			return err
		}

		if err := m.migrate(tx, new(tzip.TZIP)); err != nil {
			return err
		}

		// Re-creating primary keys after dropping part of it
		if err := tx.Exec(`ALTER TABLE big_map_states ADD PRIMARY KEY (network,contract,ptr,key_hash)`).Error; err != nil {
			return err
		}

		return tx.Exec(`ALTER TABLE token_balances ADD PRIMARY KEY (network,contract,address,token_id)`).Error
	})

}

func (m *NetworkType) migrate(tx *gorm.DB, model models.Model, indices ...string) error {
	migrator := tx.Migrator()

	logger.Info("Migrating %s....", model.GetIndex())
	columnTypes, err := migrator.ColumnTypes(model)
	if err != nil {
		return err
	}

	for i := range columnTypes {
		if columnTypes[i].Name() != networkColumn {
			continue
		}
		if columnTypes[i].DatabaseTypeName() != "text" {
			continue
		}

		if err := migrator.RenameColumn(model, networkColumn, bufColumn); err != nil {
			return err
		}
		if err := migrator.AddColumn(model, networkColumn); err != nil {
			return err
		}

		for name, network := range namesToNetwork {
			if err := tx.Table(model.GetIndex()).Where("buf_column = ?", name).Updates(map[string]interface{}{networkColumn: network}).Error; err != nil {
				return err
			}
		}

		for _, idx := range indices {
			if !migrator.HasIndex(model, idx) {
				continue
			}

			if err := migrator.DropIndex(model, idx); err != nil {
				return err
			}
		}

		if err := migrator.DropColumn(model, bufColumn); err != nil {
			return err
		}

		break
	}

	return migrator.AutoMigrate(model)
}
