package main

import (
	"sync"

	contractHandlers "github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
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
	logger.Info("%d big map diff processed        models=%d", len(bmd), len(items))

	if err := saveSearchModels(ctx.Searcher, items); err != nil {
		return err
	}

	return ctx.Storage.Save(items)
}

func initHandlers() {
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewLedger(ctx.Storage, ctx.Operations, ctx.TokenBalances, ctx.SharePath),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTokenMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Storage, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTZIP(ctx.BigMapDiffs, ctx.Blocks, ctx.Storage, ctx.TZIP, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTezosDomains(ctx.Storage, ctx.Operations, ctx.TezosDomainsContracts, ctx.SharePath),
	)
}

func parseBigMapDiff(bmd bigmapdiff.BigMapDiff) ([]models.Model, error) {
	h := metrics.New(ctx.Contracts, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TZIP, ctx.Storage)

	if err := h.SetBigMapDiffsStrings(&bmd); err != nil {
		return nil, err
	}

	items := make([]models.Model, 0)
	if len(bmd.KeyStrings) > 0 || len(bmd.ValueStrings) > 0 {
		items = append(items, &bmd)
	}

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
