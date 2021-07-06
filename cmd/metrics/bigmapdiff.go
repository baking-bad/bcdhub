package main

import (
	"sync"

	contractHandlers "github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/pkg/errors"
)

var bigMapDiffHandlers = []contractHandlers.Handler{}
var bigMapDiffHandlersInit = sync.Once{}

func getBigMapDiff(ids []int64) error {
	bigMapDiffHandlersInit.Do(initHandlers)

	bmd, err := ctx.BigMapDiffs.GetByIDs(ids...)
	if err != nil {
		return errors.Errorf("[getBigMapDiff] Find big map diff error for IDs %v: %s", ids, err)
	}

	items := make([]models.Model, 0)
	for i := range bmd {
		res, err := parseBigMapDiff(bmd[i])
		if err != nil {
			return errors.Errorf("[getBigMapDiff] Compute error message: %s", err)
		}
		items = append(items, res...)
	}

	logger.Info().Int("models", len(items)).Msgf("%2d big map diff processed", len(bmd))

	if len(items) > 0 {
		if err := ctx.Storage.Save(items); err != nil {
			return err
		}
	}

	for i := range bmd {
		if len(bmd[i].KeyStrings) > 0 || len(bmd[i].ValueStrings) > 0 {
			items = append(items, &bmd[i])
		}
	}

	if len(items) > 0 {
		return saveSearchModels(ctx.Searcher, items)
	}
	return nil
}

func initHandlers() {
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTokenMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.TokenMetadata, ctx.Storage, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTZIP(ctx.BigMapDiffs, ctx.Blocks, ctx.Storage, ctx.TZIP, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTezosDomains(ctx.Storage, ctx.Operations, ctx.TezosDomainsContracts),
	)
}

func parseBigMapDiff(bmd bigmapdiff.BigMapDiff) ([]models.Model, error) {
	items := make([]models.Model, 0)

	storageType, err := getStorageType(bmd)
	if err != nil {
		return nil, err
	}

	for i := range bigMapDiffHandlers {
		if ok, res, err := bigMapDiffHandlers[i].Do(&bmd, storageType); err != nil {
			return nil, err
		} else if ok {
			items = append(items, res...)
			break
		}
	}

	return items, nil
}
