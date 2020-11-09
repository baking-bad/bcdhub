package rollback

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

// Manager -
type Manager struct {
	e            elastic.IElastic
	messageQueue mq.IMessagePublisher
	rpc          noderpc.INode
	sharePath    string
}

// NewManager -
func NewManager(e elastic.IElastic, messageQueue mq.IMessagePublisher, rpc noderpc.INode, sharePath string) Manager {
	return Manager{
		e, messageQueue, rpc, sharePath,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(fromState models.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}
	affectedContractIDs, err := rm.getAffectedContracts(fromState.Network, fromState.Level, toLevel)
	if err != nil {
		return err
	}
	logger.Info("Rollback will affect %d contracts", len(affectedContractIDs))

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

func (rm Manager) rollbackBlocks(network string, toLevel int64) error {
	logger.Info("Deleting blocks...")
	return rm.e.DeleteByLevelAndNetwork([]string{elastic.DocBlocks}, network, toLevel)
}

func (rm Manager) rollbackOperations(network string, toLevel int64) error {
	logger.Info("Deleting operations, migrations, transfers and big map diffs...")
	return rm.e.DeleteByLevelAndNetwork([]string{elastic.DocBigMapDiff, elastic.DocBigMapActions, elastic.DocTZIP, elastic.DocMigrations, elastic.DocOperations, elastic.DocTransfers}, network, toLevel)
}

func (rm Manager) rollbackContracts(fromState models.Block, toLevel int64) error {
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
	return rm.e.DeleteByLevelAndNetwork([]string{elastic.DocContracts}, fromState.Network, toLevel)
}

func (rm Manager) getAffectedContracts(network string, fromLevel, toLevel int64) ([]string, error) {
	addresses, err := rm.e.GetAffectedContracts(network, fromLevel, toLevel)
	if err != nil {
		return nil, err
	}

	return rm.e.GetContractsIDByAddress(addresses, network)
}

func (rm Manager) getProtocolByLevel(protocols []models.Protocol, level int64) (models.Protocol, error) {
	for _, p := range protocols {
		if p.StartLevel <= level {
			return p, nil
		}
	}
	if len(protocols) == 0 {
		return models.Protocol{}, errors.Errorf("Can't find protocol for level %d", level)
	}
	return protocols[0], nil
}

func (rm Manager) removeMetadata(fromState models.Block, toLevel int64) error {
	logger.Info("Preparing metadata for removing...")
	addresses, err := rm.e.GetContractAddressesByNetworkAndLevel(fromState.Network, toLevel)
	if err != nil {
		return err
	}

	return rm.removeContractsMetadata(fromState.Network, addresses, fromState.Protocol)
}

func (rm Manager) removeContractsMetadata(network string, addresses []string, protocol string) error {
	bulkDeleteMetadata := make([]elastic.Model, 0)

	logger.Info("%d contracts will be removed", len(addresses))
	bar := progressbar.NewOptions(len(addresses), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for _, address := range addresses {
		bar.Add(1) //nolint
		bulkDeleteMetadata = append(bulkDeleteMetadata, &models.Metadata{
			ID: address,
		})

		if err := contractparser.RemoveContractFromFileSystem(address, network, protocol, rm.sharePath); err != nil {
			return err
		}
	}

	logger.Info("Removing metadata...")
	if len(bulkDeleteMetadata) > 0 {
		if err := rm.e.BulkDelete(bulkDeleteMetadata); err != nil {
			return err
		}
	}
	return nil
}

func (rm Manager) updateMetadata(network string, fromLevel, toLevel int64) error {
	logger.Info("Preparing metadata for updating...")
	var protocols []models.Protocol
	if err := rm.e.GetByNetworkWithSort(network, "start_level", "desc", &protocols); err != nil {
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
	deadSymLinks, err := rm.e.GetSymLinks(network, toLevel)
	if err != nil {
		return err
	}

	delete(deadSymLinks, rollbackProtocol.SymLink)

	logger.Info("Getting all metadata...")
	var metadata []models.Metadata
	if err := rm.e.GetAll(&metadata); err != nil {
		return err
	}

	logger.Info("Found %d metadata, will remove %v", len(metadata), deadSymLinks)
	bulkUpdateMetadata := make([]elastic.Model, len(metadata))
	for i := range metadata {
		bulkUpdateMetadata[i] = &metadata[i]
	}

	if len(bulkUpdateMetadata) > 0 {
		for symLink := range deadSymLinks {
			for i := 0; i < len(bulkUpdateMetadata); i += 1000 {
				start := i * 1000
				end := helpers.MinInt((i+1)*1000, len(bulkUpdateMetadata))
				parameterScript := fmt.Sprintf("ctx._source.parameter.remove('%s')", symLink)
				if err := rm.e.BulkRemoveField(parameterScript, bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
				storageScript := fmt.Sprintf("ctx._source.storage.remove('%s')", symLink)
				if err := rm.e.BulkRemoveField(storageScript, bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
