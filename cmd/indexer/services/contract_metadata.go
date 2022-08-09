package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/domains"
)

// ContractMetadataHandler -
type ContractMetadataHandler struct {
	*config.Context
	handler *handlers.ContractMetadata
}

// NewContractMetadataHandler -
func NewContractMetadataHandler(ctx *config.Context) *ContractMetadataHandler {
	return &ContractMetadataHandler{
		ctx,
		handlers.NewContractMetadata(ctx, ctx.Config.IPFSGateways),
	}
}

// Handle -
func (cm *ContractMetadataHandler) Handle(ctx context.Context, items []domains.BigMapDiff, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	var localWg sync.WaitGroup
	var mx sync.Mutex

	updates := make([]*contract_metadata.ContractMetadata, 0)
	for i := range items {
		localWg.Add(1)
		go func(bmd *domains.BigMapDiff) {
			defer localWg.Done()

			res, err := cm.handler.Do(ctx, bmd, nil)
			if err != nil {
				logger.Warning().Err(err).Msgf("ContractMetadata.Handle")
				return
			}
			if len(res) > 0 {
				mx.Lock()
				updates = append(updates, res...)
				mx.Unlock()
			}
		}(&items[i])
	}

	localWg.Wait()

	if len(updates) == 0 {
		return nil
	}

	logger.Info().Str("network", cm.Network.String()).Msgf("%3d contract metadata are processed", len(updates))

	return save(ctx, cm.StorageDB.DB, updates)
}

// Chunk -
func (cm *ContractMetadataHandler) Chunk(lastID int64, size int) ([]domains.BigMapDiff, error) {
	return cm.Domains.BigMapDiffs(lastID, int64(size))
}
