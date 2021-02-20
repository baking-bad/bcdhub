package rollback

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
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
	if err := rm.removeMetadata(fromState, toLevel); err != nil {
		return err
	}
	if err := rm.updateMetadata(fromState.Network, fromState.Level, toLevel); err != nil {
		return err
	}

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

func (rm Manager) removeMetadata(fromState block.Block, toLevel int64) error {
	logger.Info("Preparing metadata for removing...")
	addresses, err := rm.contractsRepo.GetAddressesByNetworkAndLevel(fromState.Network, toLevel)
	if err != nil {
		return err
	}

	return rm.removeContractsMetadata(fromState.Network, addresses, fromState.Protocol)
}

func (rm Manager) removeContractsMetadata(network string, addresses []string, protocol string) error {
	bulkDeleteMetadata := make([]models.Model, 0)

	logger.Info("%d contracts will be removed", len(addresses))
	bar := progressbar.NewOptions(len(addresses), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for _, address := range addresses {
		bar.Add(1) //nolint
		bulkDeleteMetadata = append(bulkDeleteMetadata, &schema.Schema{
			ID: address,
		})

		if err := fetch.RemoveContractFromFileSystem(address, network, protocol, rm.sharePath); err != nil {
			return err
		}
	}

	logger.Info("Removing metadata...")
	if len(bulkDeleteMetadata) > 0 {
		if err := rm.storage.BulkDelete(bulkDeleteMetadata); err != nil {
			return err
		}
	}
	return nil
}

func (rm Manager) updateMetadata(network string, fromLevel, toLevel int64) error {
	logger.Info("Preparing metadata for updating...")
	var protocols []protocol.Protocol
	if err := rm.storage.GetByNetworkWithSort(network, "start_level", "desc", &protocols); err != nil {
		return err
	}
	rollbackProtocol, err := rm.getProtocolByLevel(protocols, toLevel)
	if err != nil {
		return err
	}

	currentProtocol, err := rm.getProtocolByLevel(protocols, fromLevel)
	if err != nil {
		return err
	}

	if currentProtocol.Alias == rollbackProtocol.Alias {
		return nil
	}

	logger.Info("Rollback to %s from %s", rollbackProtocol.Hash, currentProtocol.Hash)
	deadSymLinks, err := rm.protocolsRepo.GetSymLinks(network, toLevel)
	if err != nil {
		return err
	}

	delete(deadSymLinks, rollbackProtocol.SymLink)

	logger.Info("Getting all metadata...")
	var metadata []schema.Schema
	if err := rm.storage.GetAll(&metadata); err != nil {
		return err
	}

	logger.Info("Found %d metadata, will remove %v", len(metadata), deadSymLinks)
	bulkUpdateMetadata := make([]models.Model, len(metadata))
	for i := range metadata {
		bulkUpdateMetadata[i] = &metadata[i]
	}

	if len(bulkUpdateMetadata) > 0 {
		for symLink := range deadSymLinks {
			for i := 0; i < len(bulkUpdateMetadata); i += 1000 {
				start := i * 1000
				end := helpers.MinInt((i+1)*1000, len(bulkUpdateMetadata))
				if err := rm.storage.BulkRemoveField(fmt.Sprintf("parameter.%s", symLink), bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
				if err := rm.storage.BulkRemoveField(fmt.Sprintf("storage.%s", symLink), bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
