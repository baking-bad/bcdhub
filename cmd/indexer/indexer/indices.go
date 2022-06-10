package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/go-pg/pg/v10"
)

func createStartIndices(db pg.DBI) error {
	return db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		// Blocks
		if _, err := db.Model((*block.Block)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS blocks_level_idx ON ?TableName (level)`); err != nil {
			return err
		}

		// Big map diff
		if _, err := db.Model((*bigmapdiff.BigMapDiff)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_diff_idx ON ?TableName (contract, ptr)`); err != nil {
			return err
		}

		if _, err := db.Model((*bigmapdiff.BigMapDiff)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_diff_key_hash_idx ON ?TableName (key_hash, ptr)`); err != nil {
			return err
		}

		// Big map state
		if _, err := db.Model((*bigmapdiff.BigMapState)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_state_ptr_idx ON ?TableName (ptr)`); err != nil {
			return err
		}

		if _, err := db.Model((*bigmapdiff.BigMapState)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_state_contract_idx ON ?TableName (contract)`); err != nil {
			return err
		}

		if _, err := db.Model((*bigmapdiff.BigMapState)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_state_last_update_level_idx ON ?TableName (last_update_level)`); err != nil {
			return err
		}

		// Contracts
		if _, err := db.Model((*contract.Contract)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS contracts_account_id_idx ON ?TableName (account_id)`); err != nil {
			return err
		}

		// Token balances
		if _, err := db.Model((*tokenbalance.TokenBalance)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS token_balances_by_token_idx ON ?TableName (contract, token_id)`); err != nil {
			return err
		}
		if _, err := db.Model((*tokenbalance.TokenBalance)(nil)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS token_balances_account_idx ON ?TableName (account_id)`); err != nil {
			return err
		}

		// Contract Metadata
		if _, err := db.Model(new(contract_metadata.ContractMetadata)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS contract_metadata_address_idx ON ?TableName (address)`); err != nil {
			return err
		}

		if _, err := db.Model(new(contract_metadata.ContractMetadata)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS contract_metadata_events_idx ON contract_metadata (events) where events is not null`); err != nil {
			return err
		}

		if _, err := db.Model(new(contract_metadata.ContractMetadata)).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS contract_metadata_updated_at_idx ON ?TableName (updated_at)`); err != nil {
			return err
		}

		return nil
	})
}

func (bi *BlockchainIndexer) createIndices() {
	logger.Info().Str("network", bi.Network.String()).Msg("creating database indices...")

	// Big map action
	if _, err := bi.Context.StorageDB.DB.Model((*bigmapaction.BigMapAction)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_action_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*bigmapaction.BigMapAction)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_actions_source_ptr_idx ON ?TableName (source_ptr) where source_ptr is not null
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*bigmapaction.BigMapAction)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_actions_destination_ptr_idx ON ?TableName (destination_ptr) where destination_ptr is not null
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Big map diff
	if _, err := bi.Context.StorageDB.DB.Model((*bigmapdiff.BigMapDiff)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_diff_operation_id_idx ON ?TableName (operation_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*bigmapdiff.BigMapDiff)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_diff_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*bigmapdiff.BigMapDiff)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS big_map_diff_protocol_idx ON ?TableName (protocol_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Contracts
	if _, err := bi.Context.StorageDB.DB.Model((*contract.Contract)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS contracts_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*contract.Contract)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS contracts_alpha_id_idx ON ?TableName (alpha_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*contract.Contract)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS contracts_babylon_id_idx ON ?TableName (babylon_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// DApps
	if _, err := bi.Context.StorageDB.DB.Model((*dapp.DApp)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS dapps_slug_idx ON ?TableName (slug)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Global constants
	if _, err := bi.Context.StorageDB.DB.Model((*contract.GlobalConstant)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS global_constants_address_idx ON ?TableName (address)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Migrations
	if _, err := bi.Context.StorageDB.DB.Model((*migration.Migration)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS migrations_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*migration.Migration)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS migrations_contract_id_idx ON ?TableName (contract_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Operations
	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_source_idx ON ?TableName (source_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_destination_idx ON ?TableName (destination_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_opg_idx ON ?TableName (hash, counter, content_index)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_entrypoint_idx ON ?TableName (entrypoint)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_hash_idx ON ?TableName (hash)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_opg_with_nonce_idx ON ?TableName (hash, counter, nonce)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_sort_idx ON ?TableName (level, counter, id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*operation.Operation)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS operations_status_idx ON ?TableName (status)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Token metadata
	if _, err := bi.Context.StorageDB.DB.Model((*tokenmetadata.TokenMetadata)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS token_metadata_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*tokenmetadata.TokenMetadata)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS token_metadata_name_idx ON ?TableName (name)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*tokenmetadata.TokenMetadata)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS token_metadata_creators_idx ON ?TableName USING gin (creators)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Transfers
	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_from_idx ON ?TableName ("from_id")
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_to_idx ON ?TableName ("to_id")
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_operation_id_idx ON ?TableName (operation_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_timestamp_idx ON ?TableName (timestamp)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	if _, err := bi.Context.StorageDB.DB.Model((*transfer.Transfer)(nil)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS transfers_by_token_idx ON ?TableName (contract, token_id)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}

	// Contract Metadata
	if _, err := bi.Context.StorageDB.DB.Model(new(contract_metadata.ContractMetadata)).Exec(`
		CREATE INDEX CONCURRENTLY IF NOT EXISTS tzips_level_idx ON ?TableName (level)
	`); err != nil {
		logger.Error().Err(err).Msg("can't create index")
	}
}
