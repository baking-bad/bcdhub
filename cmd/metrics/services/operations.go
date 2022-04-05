package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/search"
)

// OperationsHandler -
type OperationsHandler struct {
	*config.Context
}

// NewOperationsHandler -
func NewOperationsHandler(ctx *config.Context) *OperationsHandler {
	return &OperationsHandler{ctx}
}

// Handle -
func (oh *OperationsHandler) Handle(ctx context.Context, items []*operation.Operation, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Str("network", oh.Network.String()).Msgf("%3d operations are processed", len(items))

	return search.Save(ctx, oh.Searcher, oh.Network, items)
}

// Chunk -
func (oh *OperationsHandler) Chunk(lastID int64, size int) ([]*operation.Operation, error) {
	return getOperations(oh.StorageDB.DB, lastID, size)
}
