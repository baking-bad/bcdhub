package rollback

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/schollz/progressbar/v3"
)

// Rollback - rollback indexer state to level
func Rollback(e *elastic.Elastic, messageQueue *mq.MQ, fromState models.State, toLevel int64) error {
	if toLevel >= fromState.Level {
		return fmt.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}
	affectedContractIDs, err := getAffectedContracts(e, fromState.Network, fromState.Level, toLevel)
	if err != nil {
		return nil
	}
	logger.Info("Rollback will affect %d contracts", len(affectedContractIDs))

	if err := rollbackOperations(e, fromState.Network, toLevel); err != nil {
		return err
	}
	if err := rollbackContracts(e, fromState.Network, fromState.Level, toLevel); err != nil {
		return err
	}
	if err := rollbackState(e, fromState, toLevel); err != nil {
		return err
	}

	logger.Info("Sending to queue affected contract ids...")
	for i := range affectedContractIDs {
		if err := messageQueue.Send(mq.ChannelNew, mq.QueueRollback, affectedContractIDs[i]); err != nil {
			return err
		}
	}

	return nil
}

func rollbackState(e *elastic.Elastic, fromState models.State, toLevel int64) error {
	protocols, err := e.GetProtocolsByNetwork(fromState.Network)
	if err != nil {
		return err
	}
	rollbackProtocol, err := getProtocolByLevel(protocols, toLevel)
	if err != nil {
		return err
	}

	fromState.Level = toLevel
	fromState.Protocol = rollbackProtocol.Hash
	_, err = e.UpdateDoc(elastic.DocStates, fromState.ID, fromState)
	return err
}
func rollbackOperations(e *elastic.Elastic, network string, toLevel int64) error {
	logger.Info("Deleting operations, migrations and big map diffs...")
	return e.DeleteByLevelAndNetwork([]string{elastic.DocBigMapDiff, elastic.DocMigrations, elastic.DocOperations}, network, toLevel)
}

func rollbackContracts(e *elastic.Elastic, network string, fromLevel, toLevel int64) error {
	if err := removeMetadata(e, network, fromLevel, toLevel); err != nil {
		return err
	}
	if err := updateMetadata(e, network, fromLevel, toLevel); err != nil {
		return err
	}

	logger.Info("Deleting contracts...")
	if toLevel == 0 {
		toLevel = -1
	}
	return e.DeleteByLevelAndNetwork([]string{elastic.DocContracts}, network, toLevel)
}

func getAffectedContracts(es *elastic.Elastic, network string, fromLevel, toLevel int64) ([]string, error) {
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

func removeMetadata(e *elastic.Elastic, network string, fromLevel, toLevel int64) error {
	logger.Info("Preparing metadata for removing...")
	contracts, err := e.GetContractAddressesByNetworkAndLevel(network, toLevel)
	if err != nil {
		return err
	}

	bulkDeleteMetadata := make([]elastic.Identifiable, 0)

	arr := contracts.Array()
	bar := progressbar.NewOptions(len(arr), progressbar.OptionSetPredictTime(false))
	for _, contract := range arr {
		bar.Add(1)
		address := contract.Get("_source.address").String()
		bulkDeleteMetadata = append(bulkDeleteMetadata, models.Metadata{
			ID: address,
		})
	}
	fmt.Print("\033[2K\r")

	logger.Info("Removing metadata...")
	if len(bulkDeleteMetadata) > 0 {
		if err := e.BulkDelete(elastic.DocMetadata, bulkDeleteMetadata); err != nil {
			return err
		}
	}
	logger.Info("Removed")
	return nil
}

func updateMetadata(e *elastic.Elastic, network string, fromLevel, toLevel int64) error {
	logger.Info("Preparing metadata for updating...")
	protocols, err := e.GetProtocolsByNetwork(network)
	if err != nil {
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

	restAliases, err := e.GetSymLinksToLevel(network, toLevel)
	if err != nil {
		return err
	}

	logger.Info("Getting all metadata...")
	metadata, err := e.GetAllMetadata(map[string]interface{}{
		"network": network,
	})
	bulkUpdateMetadata := make([]elastic.Identifiable, 0)

	bar := progressbar.NewOptions(len(metadata), progressbar.OptionSetPredictTime(false))
	for _, m := range metadata {
		bar.Add(1)
		newMetadata, err := changeMetadata(m, restAliases)
		if err != nil {
			return err
		}
		bulkUpdateMetadata = append(bulkUpdateMetadata, newMetadata)
	}
	fmt.Print("\033[2K\r")

	logger.Info("Updating metadata...")
	if len(bulkUpdateMetadata) > 0 {
		if err := e.BulkUpdate(elastic.DocMetadata, bulkUpdateMetadata); err != nil {
			return err
		}
	}
	logger.Info("Updated")
	return nil
}

func changeMetadata(metadata models.Metadata, restAliases []string) (models.Metadata, error) {
	newMetadata := models.Metadata{
		ID:        metadata.ID,
		Parameter: make(map[string]string),
		Storage:   make(map[string]string),
	}
	for _, alias := range restAliases {
		p, ok := metadata.Parameter[alias]
		if !ok {
			return newMetadata, fmt.Errorf("[Rollback] Unknown parameter alias: %s", alias)
		}
		newMetadata.Parameter[alias] = p
		s, ok := metadata.Storage[alias]
		if !ok {
			return newMetadata, fmt.Errorf("[Rollback] Unknown storage alias: %s", alias)
		}
		newMetadata.Storage[alias] = s
	}
	return metadata, nil
}
