package index

import (
	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

// Indexer -
type Indexer interface {
	GetHead() (Head, error)
	GetContracts(startLevel int64) ([]Contract, error)
	GetContractOperationBlocks(startBlock int, knownContracts []models.Contract) ([]int64, error)
}
