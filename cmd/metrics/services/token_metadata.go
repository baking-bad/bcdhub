package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
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
		ctx,
		handlers.NewTokenMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.TokenMetadata, ctx.Storage, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	}
}

// Handle -
func (tm *TokenMetadataHandler) Handle(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	updates := make([]models.Model, 0)
	for i := range items {
		bmd, ok := items[i].(*bigmapdiff.BigMapDiff)
		if !ok {
			return errors.Errorf("[TokenMetadata.Handle] invalid type: expected *bigmapdiff.BigMapDiff got %T", items[i])
		}

		protocol, err := tm.CachedProtocolByID(bmd.Network, bmd.ProtocolID)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] can't get protocol by ID %d in %s: %s", bmd.ProtocolID, bmd.Network.String(), err)
		}

		storageType, err := tm.CachedStorageType(bmd.Network, bmd.Contract, protocol.SymLink)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] can't get storage type for '%s' in %s: %s", bmd.Contract, bmd.Network.String(), err)
		}

		res, err := tm.handler.Do(bmd, storageType)
		if err != nil {
			return errors.Errorf("[TokenMetadata.Handle] compute error message: %s", err)
		}

		updates = append(updates, res...)
	}

	if len(updates) == 0 {
		return nil
	}

	logger.Info("%2d token metadata are processed", len(updates))

	if err := tm.Storage.Save(updates); err != nil {
		return err
	}
	return saveSearchModels(tm.Context, updates)
}

// Chunk -
func (tm *TokenMetadataHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	var diff []bigmapdiff.BigMapDiff
	if err := getModels(tm.StorageDB.DB, models.DocBigMapDiff, lastID, size, &diff); err != nil {
		return nil, err
	}

	data := make([]models.Model, len(diff))
	for i := range diff {
		data[i] = &diff[i]
	}
	return data, nil
}
