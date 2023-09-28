package core

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/uptrace/bun"
)

func (p *Postgres) CreateIndex(ctx context.Context, name, columns string, model any) error {
	_, err := p.DB.NewCreateIndex().
		Model(model).
		IfNotExists().
		Index(name).
		ColumnExpr(columns).
		Exec(ctx)
	return err
}

func createBaseIndices(ctx context.Context, db bun.IDB) error {
	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Blocks
		if _, err := db.NewCreateIndex().
			Model((*block.Block)(nil)).
			IfNotExists().
			Index("blocks_level_idx").
			Column("level").
			Exec(ctx); err != nil {
			return err
		}

		// Big map diff
		if _, err := db.NewCreateIndex().
			Model((*bigmapdiff.BigMapDiff)(nil)).
			IfNotExists().
			Index("big_map_diff_idx").
			ColumnExpr("contract, ptr").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateIndex().
			Model((*bigmapdiff.BigMapDiff)(nil)).
			IfNotExists().
			Index("big_map_diff_key_hash_idx").
			ColumnExpr("key_hash, ptr").
			Exec(ctx); err != nil {
			return err
		}

		// Big map state
		if _, err := db.NewCreateIndex().
			Model((*bigmapdiff.BigMapState)(nil)).
			IfNotExists().
			Index("big_map_state_ptr_idx").
			Column("ptr").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateIndex().
			Model((*bigmapdiff.BigMapState)(nil)).
			IfNotExists().
			Index("big_map_state_contract_idx").
			Column("contract").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateIndex().
			Model((*bigmapdiff.BigMapState)(nil)).
			IfNotExists().
			Index("big_map_state_last_update_level_idx").
			Column("last_update_level").
			Exec(ctx); err != nil {
			return err
		}

		// Contracts
		if _, err := db.NewCreateIndex().
			Model((*contract.Contract)(nil)).
			IfNotExists().
			Index("contracts_account_id_idx").
			Column("account_id").
			Exec(ctx); err != nil {
			return err
		}

		// Scripts
		if _, err := db.NewCreateIndex().
			Model((*contract.Script)(nil)).
			IfNotExists().
			Unique().
			Index("script_hash_idx").
			Column("hash").
			Exec(ctx); err != nil {
			return err
		}

		// Operations
		if _, err := db.NewCreateIndex().
			Model((*operation.Operation)(nil)).
			IfNotExists().
			Index("operations_destination_idx").
			Column("destination_id").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateIndex().
			Model((*operation.Operation)(nil)).
			IfNotExists().
			Index("operations_status_idx").
			Column("status").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
