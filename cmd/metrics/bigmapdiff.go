package main

import (
	"sync"

	"github.com/baking-bad/bcdhub/internal/elastic"
	contractHandlers "github.com/baking-bad/bcdhub/internal/handlers"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

var bigMapDiffHandlers = []contractHandlers.Handler{}
var bigMapDiffHandlersInit = sync.Once{}

func getBigMapDiff(ids []string) error {
	bigMapDiffHandlersInit.Do(initHandlers)

	bmd := make([]models.BigMapDiff, 0)
	if err := ctx.ES.GetByIDs(&bmd, ids...); err != nil {
		return errors.Errorf("[getBigMapDiff] Find big map diff error for IDs %v: %s", ids, err)
	}

	r := result{
		Updated: make([]elastic.Model, 0),
	}
	for i := range bmd {
		if err := parseBigMapDiff(bmd[i], &r); err != nil {
			return errors.Errorf("[getBigMapDiff] Compute error message: %s", err)
		}
	}
	if err := ctx.ES.BulkUpdate(r.Updated); err != nil {
		return err
	}
	return nil
}

func initHandlers() {
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTZIP(ctx.ES, ctx.RPC, ctx.Config.IPFSGateways),
	)
	bigMapDiffHandlers = append(bigMapDiffHandlers,
		contractHandlers.NewTezosDomains(ctx.ES, ctx.Domains),
	)
}

type result struct {
	Updated []elastic.Model
}

//nolint
func parseBigMapDiff(bmd models.BigMapDiff, r *result) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if err := h.SetBigMapDiffsStrings(&bmd); err != nil {
		return err
	}
	r.Updated = append(r.Updated, &bmd)

	for i := range bigMapDiffHandlers {
		if ok, err := bigMapDiffHandlers[i].Do(&bmd); err != nil {
			return err
		} else if ok {
			break
		}
	}
	return nil
}
