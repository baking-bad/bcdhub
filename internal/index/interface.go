package index

// Indexer -
type Indexer interface {
	GetHead() (Head, error)
	GetContracts(startLevel int64) ([]Contract, error)
	GetContractOperationBlocks(startBlock int) ([]int64, error)
}
