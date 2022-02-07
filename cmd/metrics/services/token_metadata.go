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

// TokenMetadataHandler -
type TokenMetadataHandler struct {
	*config.Context
	handler *handlers.TokenMetadata
}

// NewTokenMetadataHandler -
func NewTokenMetadataHandler(ctx *config.Context) *TokenMetadataHandler {
	return &TokenMetadataHandler{
		Context: ctx,
		handler: handlers.NewTokenMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Contracts, ctx.TokenMetadata, ctx.Storage, ctx.RPC, ctx.Config.IPFSGateways),
	}
}

// Handle -
func (tm *TokenMetadataHandler) Handle(ctx context.Context, items []models.Model, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}
	var localWg sync.WaitGroup

	updates := make([]models.Model, 0)
	for i := range items {
		bmd, ok := items[i].(*domains.BigMapDiff)
		if !ok {
			return errors.Errorf("[TokenMetadata.Handle] invalid type: expected *domains.BigMapDiff got %T", items[i])
		}

		storageType, err := tm.Cache.StorageType(bmd.Network, bmd.Contract, bmd.Protocol.SymLink)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] can't get storage type for '%s' in %s: %s", bmd.Contract, bmd.Network.String(), err)
		}

		wg.Add(1)
		localWg.Add(1)
		func(wg *sync.WaitGroup) {
			defer func() {
				wg.Done()
				localWg.Done()
			}()
			res, err := tm.handler.Do(bmd, storageType)
			if err != nil {
				logger.Warning().Err(err).Msgf("TokenMetadata.Handle")
				return
			}
			updates = append(updates, res...)
		}(wg)
	}

	localWg.Wait()

	if len(updates) == 0 {
		return nil
	}

	logger.Info().Msgf("%3d token metadata are processed", len(updates))

	if err := saveSearchModels(ctx, tm.Context, updates); err != nil {
		return err
	}
	return tm.Storage.Save(ctx, updates)
}

// Chunk -
func (tm *TokenMetadataHandler) Chunk(lastID int64, size int) ([]models.Model, error) {
	diff, err := tm.Domains.BigMapDiffs(lastID, int64(size))
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(diff))
	for i := range diff {
		data[i] = &diff[i]
	}
	return data, nil
}
