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
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// Manager -
type Manager struct {
	searcher      search.Searcher
	storage       models.GeneralRepository
	transfersRepo transfer.Repository
	blockRepo     block.Repository
	bmdRepo       bigmapdiff.Repository
}

// NewManager -
func NewManager(searcher search.Searcher, storage models.GeneralRepository, blockRepo block.Repository, bmdRepo bigmapdiff.Repository, transfersRepo transfer.Repository) Manager {
	return Manager{
		searcher, storage, transfersRepo, blockRepo, bmdRepo,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(ctx context.Context, db pg.DBI, fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	for level := fromState.Level; level > toLevel; level-- {
		logger.Info().Msgf("Rollback to %d block", level)

		if _, err := rm.blockRepo.Get(fromState.Network, level); err != nil {
			if rm.storage.IsRecordNotFound(err) {
				continue
			}
			return err
		}

		err := db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			if err := rm.rollbackTokenBalances(tx, fromState.Network, level); err != nil {
				return err
			}
			if err := rm.rollbackAll(tx, fromState.Network, level); err != nil {
				return err
			}
			if err := rm.rollbackOperations(tx, fromState.Network, level); err != nil {
				return err
			}
			if err := rm.rollbackBigMapState(tx, fromState.Network, level); err != nil {
				return err
			}
			return rm.searcher.Rollback(fromState.Network.String(), toLevel)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm Manager) rollbackTokenBalances(tx pg.DBI, network types.Network, level int64) error {
	transfers, err := rm.transfersRepo.GetAll(network, level)
	if err != nil {
		return err
	}
	if len(transfers) == 0 {
		return nil
	}

	balances := make(map[string]*tokenbalance.TokenBalance)
	for i := range transfers {
		if id := transfers[i].GetFromTokenBalanceID(); id != "" {
			if update, ok := balances[id]; ok {
				update.Balance = update.Balance.Add(transfers[i].Amount)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(true, true)
				balances[id] = upd
			}
		}

		if id := transfers[i].GetToTokenBalanceID(); id != "" {
			if update, ok := balances[id]; ok {
				update.Balance = update.Balance.Sub(transfers[i].Amount)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(false, true)
				balances[id] = upd
			}
		}
	}

	for _, tb := range balances {
		if err := tb.Save(tx); err != nil {
			return err
		}
	}

	return nil
}

func (rm Manager) rollbackAll(tx pg.DBI, network types.Network, level int64) error {
	for _, index := range []models.Model{
		&block.Block{}, &contract.Contract{}, &bigmapdiff.BigMapDiff{},
		&bigmapaction.BigMapAction{}, &cm.ContractMetadata{}, &migration.Migration{},
		&transfer.Transfer{}, &tokenmetadata.TokenMetadata{},
		&global_constant.GlobalConstant{},
	} {
		if _, err := tx.Model(index).
			Where("network = ?", network).
			Where("level = ?", level).
			Delete(index); err != nil {
			return err
		}

		logger.Info().
			Str("network", network.String()).
			Str("model", index.GetIndex()).
			Msg("rollback")
	}
	return nil
}

func (rm Manager) rollbackBigMapState(tx pg.DBI, network types.Network, level int64) error {
	states, err := rm.bmdRepo.StatesChangedAfter(network, level)
	if err != nil {
		return err
	}

	for i, state := range states {
		diff, err := rm.bmdRepo.LastDiff(state.Network, state.Ptr, state.KeyHash, false)
		if err != nil {
			if rm.storage.IsRecordNotFound(err) {
				if _, err := tx.Model(&states[i]).Delete(); err != nil {
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
			valuedDiff, err := rm.bmdRepo.LastDiff(state.Network, state.Ptr, state.KeyHash, true)
			if err != nil {
				return err
			}
			states[i].Value = valuedDiff.ValueBytes()
		}

		if err := states[i].Save(tx); err != nil {
			return err
		}
	}

	return nil
}

type lastAction struct {
	Address string    `pg:"address"`
	Time    time.Time `pg:"time"`
}

func (rm Manager) rollbackOperations(tx pg.DBI, network types.Network, level int64) error {
	var ops []operation.Operation
	if err := tx.Model(&operation.Operation{}).
		Where("network = ?", network).
		Where("level = ?", level).
		Select(&ops); err != nil {
		return err
	}
	if len(ops) == 0 {
		return nil
	}

	ids := make([]int64, len(ops))
	for i := range ops {
		ids[i] = ops[i].ID
	}

	if _, err := tx.Model(&operation.Operation{}).
		WhereIn("id IN (?)", ids).
		Delete(); err != nil {
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

		if _, err := tx.Query(&actions, `select max(foo.ts) as time, foo.address from (
			(select "timestamp" as ts, source as address from operations where (network = ? and source in (?)) order by id desc limit ?)
			union all
			(select "timestamp" as ts, destination as address from operations where (network = ? and destination in (?)) order by id desc limit ?)
		) as foo
		group by address
		`, network, addresses, length, network, addresses, length); err != nil {
			return err
		}

		for i := range actions {
			count, ok := contracts[actions[i].Address]
			if !ok {
				count = 1
			}
			if _, err := tx.Exec(`update contracts set tx_count = tx_count - ?, last_action = ? where address = ?;`, count, actions[i].Time.UTC(), actions[i].Address); err != nil {
				return err
			}
		}
	}

	return nil
}
