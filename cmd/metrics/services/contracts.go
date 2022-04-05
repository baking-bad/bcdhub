package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/search"
)

// ContractsHandler -
type ContractsHandler struct {
	*config.Context
}

// NewContractsHandler -
func NewContractsHandler(ctx *config.Context) *ContractsHandler {
	return &ContractsHandler{ctx}
}

// Handle -
func (ch *ContractsHandler) Handle(ctx context.Context, items []*contract.Contract, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Str("network", ch.Network.String()).Msgf("%3d contracts are processed", len(items))

	return search.Save(ctx, ch.Searcher, ch.Network, items)
}

// Chunk -
func (ch *ContractsHandler) Chunk(lastID int64, size int) ([]*contract.Contract, error) {
	return getContracts(ch.StorageDB.DB, lastID, size)
}
