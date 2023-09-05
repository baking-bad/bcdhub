package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Indexer -
type Indexer interface {
	Start(ctx context.Context)
	Index(ctx context.Context, head noderpc.Header) error
	Rollback(ctx context.Context) error
	Close() error
}
