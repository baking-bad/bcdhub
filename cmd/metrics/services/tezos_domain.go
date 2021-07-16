package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// TezosDomainHandler -
type TezosDomainHandler struct {
	*config.Context
	handler *handlers.TezosDomain
}

// NewTezosDomainHandler -
func NewTezosDomainHandler(ctx *config.Context) *TezosDomainHandler {
	return &TezosDomainHandler{
		ctx,
		handlers.NewTezosDomains(ctx.Storage, ctx.TezosDomainsContracts),
	}
}

// Handle -
func (td *TezosDomainHandler) Handle(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	updates := make([]models.Model, 0)
	for i := range items {
		bmd, ok := items[i].(*domains.BigMapDiff)
		if !ok {
			return errors.Errorf("[TezosDomain.Handle] invalid type: expected *domains.BigMapDiff got %T", items[i])
		}

		protocol, err := td.CachedProtocolByID(bmd.BigMap.Network, bmd.ProtocolID)
		if err != nil {
			return errors.Errorf("[TezosDomain.Handle] can't get protocol by ID %d in %s: %s", bmd.ProtocolID, bmd.BigMap.Network.String(), err)
		}

		storageType, err := td.CachedStorageType(bmd.BigMap.Network, bmd.BigMap.Contract, protocol.SymLink)
		if err != nil {
			return errors.Errorf("[TezosDomain.Handle] can't get storage type for '%s' in %s: %s", bmd.BigMap.Contract, bmd.BigMap.Network.String(), err)
		}

		res, err := td.handler.Do(bmd, storageType)
		if err != nil {
			return errors.Errorf("[TezosDomain.Handle] compute error message: %s", err)
		}

		updates = append(updates, res...)
	}

	if len(updates) == 0 {
		return nil
	}

	logger.Info().Msgf("%3d tezos domains are processed", len(updates))

	if err := td.Storage.Save(updates); err != nil {
		return err
	}
	return saveSearchModels(td.Context, updates)
}

// Chunk -
func (td *TezosDomainHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	diff, err := td.Domains.BigMapDiffs(lastID, size, types.EmptyTag)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(diff))
	for i := range diff {
		data[i] = &diff[i]
	}
	return data, nil
}
