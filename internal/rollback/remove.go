package rollback

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Remove -
func Remove(es elastic.IElastic, network, appDir string) error {
	if err := removeContracts(es, network, appDir); err != nil {
		return err
	}
	return removeOthers(es, network)
}

func removeOthers(es elastic.IElastic, network string) error {
	logger.Info("Deleting general data...")
	return es.DeleteByLevelAndNetwork([]string{elastic.DocBigMapDiff, elastic.DocBigMapActions, elastic.DocMigrations, elastic.DocOperations, elastic.DocTransfers, elastic.DocBlocks, elastic.DocProtocol}, network, -1)
}

func removeContracts(es elastic.IElastic, network, appDir string) error {
	contracts, err := es.GetContracts(map[string]interface{}{
		"network": network,
	})
	if err != nil {
		return err
	}

	addresses := make([]string, len(contracts))
	for i := range contracts {
		addresses[i] = contracts[i].Address
	}

	if err := removeNetworkMetadata(es, network, addresses, appDir); err != nil {
		return err
	}
	logger.Info("Deleting contracts...")
	return es.DeleteByLevelAndNetwork([]string{elastic.DocContracts}, network, -1)
}

func removeNetworkMetadata(e elastic.IElastic, network string, addresses []string, appDir string) error {
	bulkDeleteMetadata := make([]elastic.Model, len(addresses))

	logger.Info("%d contracts will be removed", len(addresses))
	for i := range addresses {
		bulkDeleteMetadata[i] = &models.Metadata{
			ID: addresses[i],
		}
	}

	logger.Info("Removing metadata...")
	if len(bulkDeleteMetadata) > 0 {
		if err := e.BulkDelete(bulkDeleteMetadata); err != nil {
			return err
		}
	}

	logger.Info("Removing contracts from file system...")
	if err := contractparser.RemoveAllContracts(network, appDir); err != nil {
		return err
	}
	return nil
}
