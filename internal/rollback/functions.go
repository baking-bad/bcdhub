package rollback

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// Operations -
func Operations(e *elastic.Elastic, network string, level int64) error {
	logger.Info("Deleting operations, migrations and big map diffs...")
	return e.DeleteByLevelAndNetwork([]string{elastic.DocBigMapDiff, elastic.DocMigrations, elastic.DocOperations}, network, level)
}

// Contracts -
func Contracts(e *elastic.Elastic, network, protocol string, level int64) error {
	logger.Info("Getting contracts...")
	contracts, err := e.GetContractAddressesByNetworkAndLevel(network, level)
	if err != nil {
		return err
	}

	for _, contract := range contracts.Array() {
		address := contract.Get("_source.address").String()
	}

	logger.Info("Deleting contracts...")
	return e.DeleteByLevelAndNetwork([]string{elastic.DocContracts}, network, level)
}
