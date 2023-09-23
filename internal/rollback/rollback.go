package rollback

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Manager -
type Manager struct {
	rpc       noderpc.INode
	storage   models.GeneralRepository
	blockRepo block.Repository
	bmdRepo   bigmapdiff.Repository
}

// NewManager -
func NewManager(rpc noderpc.INode, storage models.GeneralRepository, blockRepo block.Repository, bmdRepo bigmapdiff.Repository) Manager {
	return Manager{
		rpc, storage, blockRepo, bmdRepo,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(ctx context.Context, db bun.IDB, network types.Network, fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	for level := fromState.Level; level > toLevel; level-- {
		logger.Info().Msgf("Rollback to %d block", level)

		if _, err := rm.blockRepo.Get(ctx, level); err != nil {
			if rm.storage.IsRecordNotFound(err) {
				continue
			}
			return err
		}

		err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if err := rm.rollbackAll(ctx, tx, level); err != nil {
				return err
			}
			if err := rm.rollbackOperations(ctx, tx, level); err != nil {
				return err
			}
			if err := rm.rollbackMigrations(ctx, tx, level); err != nil {
				return err
			}
			if err := rm.rollbackBigMapState(ctx, tx, level); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm Manager) rollbackAll(ctx context.Context, tx bun.Tx, level int64) error {
	for _, index := range []models.Model{
		&block.Block{}, &contract.Contract{}, &bigmapdiff.BigMapDiff{},
		&bigmapaction.BigMapAction{}, &contract.GlobalConstant{},
	} {
		if _, err := tx.NewDelete().
			Model(index).
			Where("level = ?", level).
			Exec(ctx); err != nil {
			return err
		}

		logger.Info().
			Str("model", index.GetIndex()).
			Msg("rollback")
	}
	return nil
}

func (rm Manager) rollbackMigrations(ctx context.Context, tx bun.Tx, level int64) error {
	logger.Info().Msg("rollback migrations...")
	query := tx.NewSelect().Model(new(contract.Contract)).Column("id").Where("level > ?", level)
	if err := tx.NewDelete().Model(new(migration.Migration)).
		Where("contract_id IN (?)", query).
		Where("level = ?", level).
		Scan(ctx); err != nil {
		return err
	}
	return nil
}

func (rm Manager) rollbackBigMapState(ctx context.Context, tx bun.Tx, level int64) error {
	logger.Info().Msg("rollback big map states...")
	states, err := rm.bmdRepo.StatesChangedAfter(ctx, level)
	if err != nil {
		return err
	}

	for i, state := range states {
		diff, err := rm.bmdRepo.LastDiff(ctx, state.Ptr, state.KeyHash, false)
		if err != nil {
			if rm.storage.IsRecordNotFound(err) {
				if _, err := tx.NewDelete().Model(&states[i]).Exec(ctx); err != nil {
					return err
				}
				continue
			}
			return err
		}
		states[i].LastUpdateLevel = diff.Level
		states[i].LastUpdateTime = diff.Timestamp
		states[i].IsRollback = true

		if len(diff.Value) > 0 {
			states[i].Value = diff.ValueBytes()
			states[i].Removed = false
		} else {
			states[i].Removed = true
			valuedDiff, err := rm.bmdRepo.LastDiff(ctx, state.Ptr, state.KeyHash, true)
			if err != nil {
				if !rm.storage.IsRecordNotFound(err) {
					return err
				}
			} else {
				states[i].Value = valuedDiff.ValueBytes()
			}
		}

		if err := states[i].Save(ctx, tx); err != nil {
			return err
		}
	}

	return nil
}

type lastAction struct {
	Address string    `bun:"address"`
	Time    time.Time `bun:"time"`
}

func (rm Manager) rollbackOperations(ctx context.Context, tx bun.Tx, level int64) error {
	logger.Info().Msg("rollback operations...")
	var ops []operation.Operation
	if err := tx.NewSelect().Model(&ops).
		Where("level = ?", level).
		Scan(ctx); err != nil {
		return err
	}
	if len(ops) == 0 {
		return nil
	}

	ids := make([]int64, len(ops))
	for i := range ops {
		ids[i] = ops[i].ID
	}

	if _, err := tx.NewDelete().Model(&operation.Operation{}).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx); err != nil {
		return err
	}

	contracts := make(map[string]uint64)
	for i := range ops {
		if ops[i].IsOrigination() {
			continue
		}
		if ops[i].Destination.Type == types.AccountTypeContract {
			if _, ok := contracts[ops[i].Destination.Address]; !ok {
				contracts[ops[i].Destination.Address] = 1
			} else {
				contracts[ops[i].Destination.Address] += 1
			}
		}
		if ops[i].Source.Type == types.AccountTypeContract {
			if _, ok := contracts[ops[i].Source.Address]; !ok {
				contracts[ops[i].Source.Address] = 1
			} else {
				contracts[ops[i].Source.Address] += 1
			}
		}
	}

	if len(contracts) > 0 {
		addresses := make([]string, 0, len(contracts))
		for address := range contracts {
			addresses = append(addresses, address)
		}
		length := len(addresses) * 10

		var actions []lastAction

		if _, err := tx.NewRaw(`select max(foo.ts) as time, foo.address from (
			(select "timestamp" as ts, source as address from operations where source in (?) order by id desc limit ?)
			union all
			(select "timestamp" as ts, destination as address from operations where destination in (?) order by id desc limit ?)
		) as foo
		group by address
		`, addresses, length, addresses, length).Exec(ctx, &actions); err != nil {
			return err
		}

		for i := range actions {
			count, ok := contracts[actions[i].Address]
			if !ok {
				count = 1
			}

			_, err := tx.NewUpdate().Model((*contract.Contract)(nil)).
				Where("address = ?", actions[i].Address).
				Set("tx_count = tx_count - ?", count).
				Set("last_action = ?", actions[i].Time).
				Exec(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
