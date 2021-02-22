package rollback

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
)

// Manager -
type Manager struct {
	storage       models.GeneralRepository
	contractsRepo contract.Repository
	operationRepo operation.Repository
	transfersRepo transfer.Repository
	tbRepo        tokenbalance.Repository
	protocolsRepo protocol.Repository
	messageQueue  mq.IMessagePublisher
	rpc           noderpc.INode
	sharePath     string
}

// NewManager -
func NewManager(storage models.GeneralRepository, contractsRepo contract.Repository, operationRepo operation.Repository, transfersRepo transfer.Repository, tbRepo tokenbalance.Repository, protocolsRepo protocol.Repository, messageQueue mq.IMessagePublisher, rpc noderpc.INode, sharePath string) Manager {
	return Manager{
		storage, contractsRepo, operationRepo, transfersRepo, tbRepo, protocolsRepo, messageQueue, rpc, sharePath,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}
	affectedContractIDs, err := rm.getAffectedContracts(fromState.Network, fromState.Level, toLevel)
	if err != nil {
		return err
	}
	logger.Info("Rollback will affect %d contracts", len(affectedContractIDs))

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

	time.Sleep(time.Second) // Golden hack: Waiting while elastic remove records
	logger.Info("Sending to queue affected contract ids...")
	for i := range affectedContractIDs {
		if err := rm.messageQueue.SendRaw(mq.QueueRecalc, []byte(affectedContractIDs[i])); err != nil {
			return err
		}
	}

	return nil
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
	updates := make([]*tokenbalance.TokenBalance, 0)
	for i := range transfers {

		if id := transfers[i].GetFromTokenBalanceID(); id != "" {
			if update, ok := exists[id]; ok {
				update.Value.Add(update.Value, transfers[i].AmountBigInt)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(true, true)
				updates = append(updates, upd)
				exists[id] = upd
			}
		}

		if id := transfers[i].GetToTokenBalanceID(); id != "" {
			if update, ok := exists[id]; ok {
				update.Value.Sub(update.Value, transfers[i].AmountBigInt)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(false, true)
				updates = append(updates, upd)
				exists[id] = upd
			}
		}
	}

	return rm.tbRepo.Update(updates)
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

func (rm Manager) getAffectedContracts(network string, fromLevel, toLevel int64) ([]string, error) {
	addresses, err := rm.operationRepo.GetParticipatingContracts(network, fromLevel, toLevel)
	if err != nil {
		return nil, err
	}

	return rm.contractsRepo.GetIDsByAddresses(addresses, network)
}

func (rm Manager) getProtocolByLevel(protocols []protocol.Protocol, level int64) (protocol.Protocol, error) {
	for _, p := range protocols {
		if p.StartLevel <= level {
			return p, nil
		}
	}
	if len(protocols) == 0 {
		return protocol.Protocol{}, errors.Errorf("Can't find protocol for level %d", level)
	}
	return protocols[0], nil
}
