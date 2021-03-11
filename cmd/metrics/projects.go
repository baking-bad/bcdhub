package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func getProject(ids []string) error {
	contracts := make([]contract.Contract, 0)
	if err := ctx.Storage.GetByIDs(&contracts, ids...); err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := parseProject(contracts[i]); err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}
	}
	logger.Info("%d contracts are pulled to projects", len(contracts))
	return nil
}

func parseProject(contract contract.Contract) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return errors.Errorf("[parseContract] Error during set contract projectID: %s", err)
		}
	}
	return ctx.Storage.UpdateFields(models.DocContracts, contract.GetID(), contract, "ProjectID")
}
