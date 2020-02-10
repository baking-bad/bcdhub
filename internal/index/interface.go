package index

// Indexer -
type Indexer interface {
	GetHead() (Head, error)
	GetContracts(startLevel int64) ([]Contract, error)
	GetContractOperationBlocks(startBlock int, endBlock int, knownContracts map[string]struct{}, spendable map[string]struct{}) ([]int64, error)
}
