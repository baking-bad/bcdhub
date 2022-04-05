package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/search"
)

// BigMapDiffHandler -
type BigMapDiffHandler struct {
	*config.Context
}

// NewBigMapDiffHandler -
func NewBigMapDiffHandler(ctx *config.Context) *BigMapDiffHandler {
	return &BigMapDiffHandler{ctx}
}

// Handle -
func (bmh *BigMapDiffHandler) Handle(ctx context.Context, items []*bigmapdiff.BigMapDiff, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Str("network", bmh.Network.String()).Msgf("%3d big map diffs are processed", len(items))

	return search.Save(ctx, bmh.Searcher, bmh.Network, items)
}

// Chunk -
func (bmh *BigMapDiffHandler) Chunk(lastID int64, size int) ([]*bigmapdiff.BigMapDiff, error) {
	return getDiffs(bmh.StorageDB.DB, lastID, size)
}
