package indexer

import (
	"context"
	"sync"
)

// Indexer -
type Indexer interface {
	Sync(ctx context.Context, wg *sync.WaitGroup)
	Index(ctx context.Context, levels []int64) error
	Rollback(ctx context.Context) error
}
