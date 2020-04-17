package index

// Indexer -
type Indexer interface {
	GetHead() (Head, error)
	GetContracts(startLevel int64) ([]Contract, error)
	GetContractOperationBlocks(startBlock int64, endBlock int64) ([]int64, error)
	GetProtocols() ([]Protocol, error)
}
