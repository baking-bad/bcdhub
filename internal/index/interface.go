package index

// Indexer -
type Indexer interface {
	GetHead() (Head, error)
	GetContracts(startLevel int64) ([]Contract, error)
}
