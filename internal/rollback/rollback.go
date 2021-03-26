package rollback

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/pkg/errors"
)

// Manager -
type Manager struct {
	searcher      search.Searcher
	storage       models.GeneralRepository
	transfersRepo transfer.Repository
	protocolsRepo protocol.Repository
}

// NewManager -
func NewManager(searcher search.Searcher, storage models.GeneralRepository, transfersRepo transfer.Repository, protocolsRepo protocol.Repository) Manager {
	return Manager{
		searcher, storage, transfersRepo, protocolsRepo,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	if err := rm.rollbackTokenBalances(fromState.Network, toLevel); err != nil {
		return err
	}
	if err := rm.rollbackOperations(fromState.Network, toLevel); err != nil {
		return err
	}
	if err := rm.rollbackContracts(fromState, toLevel); err != nil {
		return err
	}
	if err := rm.rollbackBlocks(fromState.Network, toLevel); err != nil {
		return err
	}

	return rm.searcher.Rollback(fromState.Network, toLevel)
}

func (rm Manager) rollbackTokenBalances(network string, toLevel int64) error {
	transfers, err := rm.transfersRepo.GetAll(network, toLevel)
	if err != nil {
		return err
	}
	if len(transfers) == 0 {
		return nil
	}

	exists := make(map[string]*tokenbalance.TokenBalance)
	updates := make([]models.Model, 0)
	for i := range transfers {
		if id := transfers[i].GetFromTokenBalanceID(); id != "" {
			if update, ok := exists[id]; ok {
				update.Balance += transfers[i].Amount
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(true, true)
				updates = append(updates, upd)
				exists[id] = upd
			}
		}

		if id := transfers[i].GetToTokenBalanceID(); id != "" {
			if update, ok := exists[id]; ok {
				update.Balance -= transfers[i].Amount
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(false, true)
				updates = append(updates, upd)
				exists[id] = upd
			}
		}
	}

	return rm.storage.Save(updates)
}

func (rm Manager) rollbackBlocks(network string, toLevel int64) error {
	logger.Info("Deleting blocks...")
	return rm.storage.DeleteByLevelAndNetwork([]string{models.DocBlocks}, network, toLevel)
}

func (rm Manager) rollbackOperations(network string, toLevel int64) error {
	logger.Info("Deleting operations, migrations, transfers and big map diffs...")
	return rm.storage.DeleteByLevelAndNetwork([]string{models.DocBigMapDiff, models.DocBigMapActions, models.DocTZIP, models.DocMigrations, models.DocOperations, models.DocTransfers, models.DocTokenMetadata}, network, toLevel)
}

func (rm Manager) rollbackContracts(fromState block.Block, toLevel int64) error {
	logger.Info("Deleting contracts...")
	if toLevel == 0 {
		toLevel = -1
	}
	return rm.storage.DeleteByLevelAndNetwork([]string{models.DocContracts}, fromState.Network, toLevel)
}
