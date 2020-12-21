package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

func recalculateAll(ids []string) error {
	contracts := make([]models.Contract, 0)
	if err := ctx.ES.GetByIDs(&contracts, ids...); err != nil {
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

func recalc(contract models.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if _, err := h.SetContractAlias(&contract); err != nil {
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

	return ctx.ES.UpdateDoc(&contract)
}
