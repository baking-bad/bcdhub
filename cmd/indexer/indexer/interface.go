package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Indexer -
type Indexer interface {
	// Start launches indexing. It is allowed to return before indexing has stopped:
	// an implementation may own its loop through the worker group it was constructed
	// with, so callers must wait on that group rather than on Start to know that the
	// indexer is done. Start returns early on its own only when indexing cannot be
	// resumed (see errDeepReorg); otherwise it stops when ctx is cancelled.
	Start(ctx context.Context)
	Index(ctx context.Context, head noderpc.Header) error
	Rollback(ctx context.Context) error
	Close() error
}
