package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/tokens"
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
		handler: handlers.NewTokenMetadata(ctx, ctx.Config.IPFSGateways),
	}
}

// Handle -
func (tm *TokenMetadataHandler) Handle(ctx context.Context, items []domains.BigMapDiff, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}
	var localWg sync.WaitGroup
	var mx sync.Mutex

	updates := make([]*tokenmetadata.TokenMetadata, 0)
	for i := range items {
		storageTypeBytes, err := tm.Cache.StorageTypeBytes(items[i].Contract, items[i].Protocol.SymLink)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] can't get storage type for '%s' in %s: %s", items[i].Contract, tm.Network.String(), err)
		}

		storageType, err := ast.NewTypedAstFromBytes(storageTypeBytes)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] can't parse storage type for '%s' in %s: %s", items[i].Contract, tm.Network.String(), err)
		}

		if node := storageType.FindByName(tokens.TokenMetadataStorageKey, false); node == nil {
			continue
		}

		localWg.Add(1)
		go func(bmd *domains.BigMapDiff) {
			defer localWg.Done()

			res, err := tm.handler.Do(ctx, bmd, storageType)
			if err != nil {
				logger.Warning().Err(err).Msgf("TokenMetadata.Handle")
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

	logger.Info().Str("network", tm.Network.String()).Msgf("%3d token metadata are processed", len(updates))

	return save(ctx, tm.StorageDB.DB, updates)
}

// Chunk -
func (tm *TokenMetadataHandler) Chunk(lastID int64, size int) ([]domains.BigMapDiff, error) {
	return tm.Domains.BigMapDiffs(lastID, int64(size))
}
