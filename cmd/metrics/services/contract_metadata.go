package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/pkg/errors"
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
		handlers.NewContractMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Contracts, ctx.Storage, ctx.ContractMetadata, ctx.RPC, ctx.Config.IPFSGateways),
	}
}

// Handle -
func (cm *ContractMetadataHandler) Handle(ctx context.Context, items []models.Model, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	var localWg sync.WaitGroup

	updates := make([]models.Model, 0)
	for i := range items {
		bmd, ok := items[i].(*domains.BigMapDiff)
		if !ok {
			return errors.Errorf("[ContractMetadata.Handle] invalid type: expected *bigmapdiff.BigMapDiff got %T", items[i])
		}

		storageType, err := cm.Cache.StorageType(bmd.Network, bmd.Contract, bmd.Protocol.SymLink)
		if err != nil {
			return errors.Errorf("[ContractMetadata.Handle] can't get storage type for '%s' in %s: %s", bmd.Contract, bmd.Network.String(), err)
		}

		wg.Add(1)
		localWg.Add(1)
		func() {
			defer func() {
				wg.Done()
				localWg.Done()
			}()
			res, err := cm.handler.Do(bmd, storageType)
			if err != nil {
				logger.Warning().Err(err).Msgf("ContractMetadata.Handle")
				return
			}
			updates = append(updates, res...)
		}()
	}

	localWg.Wait()

	if len(updates) == 0 {
		return nil
	}

	logger.Info().Msgf("%3d contract metadata are processed", len(updates))

	if err := saveSearchModels(ctx, cm.Context, updates); err != nil {
		return err
	}

	return cm.Storage.Save(ctx, updates)
}

// Chunk -
func (cm *ContractMetadataHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	diff, err := cm.Domains.BigMapDiffs(lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(diff))
	for i := range diff {
		data[i] = &diff[i]
	}
	return data, nil
}
