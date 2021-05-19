package rollback

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Manager -
type Manager struct {
	searcher      search.Searcher
	storage       models.GeneralRepository
	transfersRepo transfer.Repository
	bmdRepo       bigmapdiff.Repository
}

// NewManager -
func NewManager(searcher search.Searcher, storage models.GeneralRepository, bmdRepo bigmapdiff.Repository, transfersRepo transfer.Repository) Manager {
	return Manager{
		searcher, storage, transfersRepo, bmdRepo,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(db *gorm.DB, fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	for level := fromState.Level - 1; level >= toLevel; level-- {
		logger.Info("Rollback to %d block", level)
		err := db.Transaction(func(tx *gorm.DB) error {

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

func (rm Manager) rollbackTokenBalances(tx *gorm.DB, network types.Network, toLevel int64) error {
	transfers, err := rm.transfersRepo.GetAll(network, toLevel)
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
				update.Value.Add(update.Value, transfers[i].Value)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(true, true)
				balances[id] = upd
			}
		}

		if id := transfers[i].GetToTokenBalanceID(); id != "" {
			if update, ok := balances[id]; ok {
				update.Value.Sub(update.Value, transfers[i].Value)
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

func (rm Manager) rollbackAll(tx *gorm.DB, network types.Network, toLevel int64) error {
	for _, index := range []models.Model{
		&block.Block{}, &contract.Contract{}, &bigmapdiff.BigMapDiff{},
		&bigmapaction.BigMapAction{}, &tzip.TZIP{}, &migration.Migration{},
		&transfer.Transfer{}, &tokenmetadata.TokenMetadata{},
	} {
		if err := tx.Unscoped().
			Where("network = ?", network).
			Where("level > ?", toLevel).
			Delete(index).
			Error; err != nil {
			return err
		}
	}
	return nil
}

func (rm Manager) rollbackBigMapState(tx *gorm.DB, network types.Network, toLevel int64) error {
	states, err := rm.bmdRepo.StatesChangedAfter(network, toLevel)
	if err != nil {
		return err
	}

	for i, state := range states {
		diff, err := rm.bmdRepo.LastDiff(state.Network, state.Ptr, state.KeyHash, false)
		if err != nil {
			if rm.storage.IsRecordNotFound(err) {
				if err := tx.Delete(&states[i]).Error; err != nil {
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
	Address string    `gorm:"address"`
	Time    time.Time `gorm:"time"`
}

func (rm Manager) rollbackOperations(tx *gorm.DB, network types.Network, toLevel int64) error {
	var ops []operation.Operation
	if err := tx.Model(&operation.Operation{}).
		Where("network = ?", network).
		Where("level > ?", toLevel).
		Find(&ops).Error; err != nil {
		return err
	}

	if err := tx.Unscoped().
		Where("network = ?", network).
		Where("level > ?", toLevel).
		Delete(&operation.Operation{}).
		Error; err != nil {
		return err
	}

	contracts := make(map[string]uint64)
	for i := range ops {
		if ops[i].IsOrigination() {
			continue
		}
		if bcd.IsContract(ops[i].Destination) {
			if _, ok := contracts[ops[i].Destination]; !ok {
				contracts[ops[i].Destination] = 1
			} else {
				contracts[ops[i].Destination] += 1
			}
		}
		if bcd.IsContract(ops[i].Source) {
			if _, ok := contracts[ops[i].Source]; !ok {
				contracts[ops[i].Source] = 1
			} else {
				contracts[ops[i].Source] += 1
			}
		}
	}

	if len(contracts) > 0 {
		addresses := make([]string, 0, len(contracts))
		for address := range contracts {
			addresses = append(addresses, address)
		}

		var actions []lastAction

		if err := tx.Raw(`select max(foo.ts) as time, foo.address from (
			(select "timestamp" as ts, source as address from operations where (network = @network and source in @contracts) order by id desc limit @length)
			union all
			(select "timestamp" as ts, destination as address from operations where (network = @network and destination in @contracts) order by id desc limit @length)
		) as foo
		group by address
		`, map[string]interface{}{
			"network":   network,
			"contracts": addresses,
			"length":    len(addresses) * 10,
		}).Scan(&actions).Error; err != nil {
			return err
		}

		for i := range actions {
			count, ok := contracts[actions[i].Address]
			if !ok {
				count = 1
			}
			if err := tx.Exec(`update contracts set tx_count = tx_count - ?, last_action = ? where address = ?;`, count, actions[i].Time.UTC(), actions[i].Address).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
