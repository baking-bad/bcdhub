package indexer

import (
	"sync"
)

// Indexer -
type Indexer interface {
	Sync(wg *sync.WaitGroup) error
	Stop()
	Index(levels []int64) error
	Rollback() error
}
