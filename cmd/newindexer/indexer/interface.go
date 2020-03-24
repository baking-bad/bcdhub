package indexer

import "sync"

// Indexer -
type Indexer interface {
	Sync(wg *sync.WaitGroup) error
	Stop()
}
