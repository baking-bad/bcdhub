package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func getContract(ids []int64) error {
	contracts, err := ctx.Contracts.GetByIDs(ids...)
	if err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	updates := make([]models.Model, 0)
	for i := range contracts {
		res, err := parseContract(&contracts[i], contracts[:i])
		if err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}

		updates = append(updates, res...)
	}

	logger.Info("Metrics of %d contracts are computed", len(contracts))

	if err := saveSearchModels(ctx.Searcher, updates); err != nil {
		return err
	}

	return ctx.Storage.Save(updates)
}

func parseContract(contract *contract.Contract, chunk []contract.Contract) ([]models.Model, error) {
	if contract.ProjectID != "" {
		return nil, nil
	}

	h := metrics.New(ctx.Contracts, ctx.Blocks, ctx.Operations, ctx.TokenBalances, ctx.TZIP, ctx.Storage)
	if err := h.SetContractProjectID(contract, chunk); err != nil {
		return nil, errors.Errorf("[parseContract] Error during set contract projectID: %s", err)
	}

	return []models.Model{contract}, nil
}
