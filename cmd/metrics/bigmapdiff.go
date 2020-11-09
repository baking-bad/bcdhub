package main

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/pkg/errors"
)

func getBigMapDiff(ids []string) error {
	bmd := make([]models.BigMapDiff, 0)
	if err := ctx.ES.GetByIDs(&bmd, ids...); err != nil {
		return errors.Errorf("[getBigMapDiff] Find big map diff error for IDs %v: %s", ids, err)
	}

	r := result{
		Updated: make([]elastic.Model, 0),
		New:     make([]elastic.Model, 0),
	}
	for i := range bmd {
		if err := parseBigMapDiff(bmd[i], &r); err != nil {
			return errors.Errorf("[getBigMapDiff] Compute error message: %s", err)
		}
	}
	if err := ctx.ES.BulkInsert(r.New); err != nil {
		return err
	}
	if err := ctx.ES.BulkUpdate(r.Updated); err != nil {
		return err
	}
	return nil
}

type result struct {
	Updated []elastic.Model
	New     []elastic.Model
}

//nolint
func parseBigMapDiff(bmd models.BigMapDiff, r *result) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if err := h.SetBigMapDiffsStrings(&bmd); err != nil {
		return err
	}
	r.Updated = append(r.Updated, &bmd)

	switch bmd.KeyHash {
	case tzip.EmptyStringKey:
		newModels, err := tzipHandler(bmd)
		if err != nil {
			return err
		}
		r.New = append(r.New, newModels...)
	}
	return nil
}

func tzipHandler(bmd models.BigMapDiff) ([]elastic.Model, error) {
	rpc, err := ctx.GetRPC(bmd.Network)
	if err != nil {
		return nil, err
	}
	tzipParser := tzip.NewParser(ctx.ES, rpc, tzip.ParserConfig{
		IPFSGateways: ctx.Config.IPFSGateways,
	})

	model, err := tzipParser.Parse(tzip.ParseContext{
		BigMapDiff: bmd,
	})
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, nil
	}

	logger.With(&bmd).Info("Big map diff with TZIP is processed")
	return []elastic.Model{model}, nil
}
