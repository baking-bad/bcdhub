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

func getBigMapDiff(ids []string) error {
	bigMapDiffHandlersInit.Do(initHandlers)

	bmd := make([]bigmapdiff.BigMapDiff, 0)
	if err := ctx.Storage.GetByIDs(&bmd, ids...); err != nil {
		return errors.Errorf("[getBigMapDiff] Find big map diff error for IDs %v: %s", ids, err)
	}

	r := result{
		Updated:  make([]models.Model, 0),
		Inserted: make([]models.Model, 0),
	}
	for i := range bmd {
		if err := parseBigMapDiff(bmd[i], &r); err != nil {
			return errors.Errorf("[getBigMapDiff] Compute error message: %s", err)
		}
	}
	logger.Info("%d big map diff processed\t\tnew=%d updated=%d", len(bmd), len(r.Inserted), len(r.Updated))
	if err := ctx.Storage.BulkUpdate(r.Updated); err != nil {
		return err
	}
	return ctx.Storage.BulkInsert(r.Inserted)
}

func initHandlers() {
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTZIP(ctx.BigMapDiffs, ctx.Blocks, ctx.Storage, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTezosDomains(ctx.Storage, ctx.Domains, ctx.SharePath),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTokenMetadata(ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Storage, ctx.RPC, ctx.SharePath, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewLedger(ctx.Storage, ctx.TokenBalances, ctx.SharePath),
	)
}

type result struct {
	Updated  []models.Model
	Inserted []models.Model
}

//nolint
func parseBigMapDiff(bmd bigmapdiff.BigMapDiff, r *result) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	if err := h.SetBigMapDiffsStrings(&bmd); err != nil {
		return err
	}
	r.Updated = append(r.Updated, &bmd)

	for i := range bigMapDiffHandlers {
		if ok, res, err := bigMapDiffHandlers[i].Do(&bmd); err != nil {
			return err
		} else if ok {
			r.Inserted = append(r.Inserted, res...)
			break
		}
	}

	return nil
}
