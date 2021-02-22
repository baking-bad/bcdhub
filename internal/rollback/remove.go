package rollback

import (
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Remove -
func Remove(storage models.GeneralRepository, contractsRepo contract.Repository, network, appDir string) error {
	if err := removeContracts(storage, contractsRepo, network, appDir); err != nil {
		return err
	}
	return removeOthers(storage, network)
}

func removeOthers(storage models.GeneralRepository, network string) error {
	logger.Info("Deleting general data...")
	return storage.DeleteByLevelAndNetwork([]string{models.DocBigMapDiff, models.DocBigMapActions, models.DocMigrations, models.DocOperations, models.DocTransfers, models.DocBlocks, models.DocProtocol}, network, -1)
}

func removeContracts(storage models.GeneralRepository, contractsRepo contract.Repository, network, appDir string) error {
	contracts, err := contractsRepo.GetMany(map[string]interface{}{
		"network": network,
	})
	if err != nil {
		return err
	}

	addresses := make([]string, len(contracts))
	for i := range contracts {
		addresses[i] = contracts[i].Address
	}

	if err := removeNetworkMetadata(storage, network, addresses, appDir); err != nil {
		return err
	}
	logger.Info("Deleting contracts...")
	return storage.DeleteByLevelAndNetwork([]string{models.DocContracts}, network, -1)
}

func removeNetworkMetadata(storage models.GeneralRepository, network string, addresses []string, appDir string) error {
	logger.Info("Removing contracts from file system...")
	if err := fetch.RemoveAllContracts(network, appDir); err != nil {
		return err
	}
	return nil
}
