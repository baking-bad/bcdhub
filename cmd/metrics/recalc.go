package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
)

func recalculateAll(ids []string) error {
	contracts := make([]contract.Contract, 0)
	if err := ctx.Storage.GetByIDs(&contracts, ids...); err != nil {
		return errors.Errorf("[recalculateAll] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := recalc(contracts[i]); err != nil {
			return errors.Errorf("[recalculateAll] Compute error message: %s", err)
		}
		logger.With(&contracts[i]).Info("Contract metrics are recalculated")
	}

	return nil
}

func recalc(contract contract.Contract) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	aliases, err := getAliases(contract.Network)
	if err != nil {
		return err
	}

	if _, err := h.SetContractAlias(&contract, aliases); err != nil {
		return err
	}

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return errors.Errorf("[recalc] Error during set contract projectID: %s", err)
		}
	}

	if err := h.UpdateContractStats(&contract); err != nil {
		return errors.Errorf("[recalc] Compute contract stats error message: %s", err)
	}

	return ctx.Storage.UpdateDoc(&contract)
}
