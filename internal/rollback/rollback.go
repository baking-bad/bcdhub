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
	"github.com/schollz/progressbar/v3"
)

// Rollback - rollback indexer state to level
func Rollback(e elastic.IElastic, messageQueue *mq.MQ, appDir string, fromState models.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return fmt.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}
	affectedContractIDs, err := getAffectedContracts(e, fromState.Network, fromState.Level, toLevel)
	if err != nil {
		return err
	}
	logger.Info("Rollback will affect %d contracts", len(affectedContractIDs))

	if err := rollbackOperations(e, fromState.Network, toLevel); err != nil {
		return err
	}
	if err := rollbackContracts(e, fromState, toLevel, appDir); err != nil {
		return err
	}
	if err := rollbackBlocks(e, fromState.Network, toLevel); err != nil {
		return err
	}

	time.Sleep(time.Second) // Golden hack: Waiting while elastic remove records
	logger.Info("Sending to queue affected contract ids...")
	for i := range affectedContractIDs {
		if err := messageQueue.SendToQueue(mq.ChannelNew, mq.QueueRecalc, affectedContractIDs[i]); err != nil {
			return err
		}
	}

	return nil
}

func rollbackBlocks(e elastic.IElastic, network string, toLevel int64) error {
	logger.Info("Deleting blocks...")
	return e.DeleteByLevelAndNetwork([]string{elastic.DocBlocks}, network, toLevel)
}

func rollbackOperations(e elastic.IElastic, network string, toLevel int64) error {
	logger.Info("Deleting operations, migrations and big map diffs...")
	return e.DeleteByLevelAndNetwork([]string{elastic.DocBigMapDiff, elastic.DocBigMapActions, elastic.DocMigrations, elastic.DocOperations, elastic.DocTransfers}, network, toLevel)
}

func rollbackContracts(e elastic.IElastic, fromState models.Block, toLevel int64, appDir string) error {
	if err := removeMetadata(e, fromState, toLevel, appDir); err != nil {
		return err
	}
	if err := updateMetadata(e, fromState.Network, fromState.Level, toLevel); err != nil {
		return err
	}

	logger.Info("Deleting contracts...")
	if toLevel == 0 {
		toLevel = -1
	}
	return e.DeleteByLevelAndNetwork([]string{elastic.DocContracts}, fromState.Network, toLevel)
}

func getAffectedContracts(es elastic.IElastic, network string, fromLevel, toLevel int64) ([]string, error) {
	addresses, err := es.GetAffectedContracts(network, fromLevel, toLevel)
	if err != nil {
		return nil, err
	}

	return es.GetContractsIDByAddress(addresses, network)
}

func getProtocolByLevel(protocols []models.Protocol, level int64) (models.Protocol, error) {
	for _, p := range protocols {
		if p.StartLevel <= level {
			return p, nil
		}
	}
	if len(protocols) == 0 {
		return models.Protocol{}, fmt.Errorf("Can't find protocol for level %d", level)
	}
	return protocols[0], nil
}

func removeMetadata(e elastic.IElastic, fromState models.Block, toLevel int64, appDir string) error {
	logger.Info("Preparing metadata for removing...")
	contracts, err := e.GetContractAddressesByNetworkAndLevel(fromState.Network, toLevel)
	if err != nil {
		return err
	}

	bulkDeleteMetadata := make([]elastic.Model, 0)

	arr := contracts.Array()
	logger.Info("%d contracts will be removed", len(arr))
	bar := progressbar.NewOptions(len(arr), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for _, contract := range arr {
		bar.Add(1) //nolint
		address := contract.Get("_source.address").String()
		bulkDeleteMetadata = append(bulkDeleteMetadata, &models.Metadata{
			ID: address,
		})

		if err := contractparser.RemoveContractFromFileSystem(address, fromState.Network, fromState.Protocol, appDir); err != nil {
			return err
		}
	}

	logger.Info("Removing metadata...")
	if len(bulkDeleteMetadata) > 0 {
		if err := e.BulkDelete(bulkDeleteMetadata); err != nil {
			return err
		}
	}
	return nil
}

func updateMetadata(e elastic.IElastic, network string, fromLevel, toLevel int64) error {
	logger.Info("Preparing metadata for updating...")
	var protocols []models.Protocol
	if err := e.GetByNetworkWithSort(network, "start_level", "desc", &protocols); err != nil {
		return err
	}
	rollbackProtocol, err := getProtocolByLevel(protocols, toLevel)
	if err != nil {
		return err
	}

	currentProtocol, err := getProtocolByLevel(protocols, fromLevel)
	if err != nil {
		return err
	}

	if currentProtocol.Alias == rollbackProtocol.Alias {
		return nil
	}

	logger.Info("Rollback to %s from %s", rollbackProtocol.Hash, currentProtocol.Hash)
	deadSymLinks, err := e.GetSymLinks(network, toLevel)
	if err != nil {
		return err
	}

	delete(deadSymLinks, rollbackProtocol.SymLink)

	logger.Info("Getting all metadata...")
	var metadata []models.Metadata
	if err := e.GetAll(&metadata); err != nil {
		return err
	}

	logger.Info("Found %d metadata, will remove %v", len(metadata), deadSymLinks)
	bulkUpdateMetadata := make([]elastic.Model, 0)
	for _, m := range metadata {
		bulkUpdateMetadata = append(bulkUpdateMetadata, &m)
	}

	if len(bulkUpdateMetadata) > 0 {
		for symLink := range deadSymLinks {
			for i := 0; i < len(bulkUpdateMetadata); i += 1000 {
				start := i * 1000
				end := helpers.MinInt((i+1)*1000, len(bulkUpdateMetadata))
				parameterScript := fmt.Sprintf("ctx._source.parameter.remove('%s')", symLink)
				if err := e.BulkRemoveField(parameterScript, bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
				storageScript := fmt.Sprintf("ctx._source.storage.remove('%s')", symLink)
				if err := e.BulkRemoveField(storageScript, bulkUpdateMetadata[start:end]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
