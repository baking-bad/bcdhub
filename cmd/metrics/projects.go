package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
)

func getProject(ids []string) error {
	contracts := make([]models.Contract, 0)
	if err := ctx.ES.GetByIDs(&contracts, ids...); err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := parseProject(contracts[i]); err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}
	}
	return nil
}

func parseProject(contract models.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return errors.Errorf("[parseContract] Error during set contract projectID: %s", err)
		}
	}
	logger.With(&contract).Infof("Contract is pulled to project %s", contract.ProjectID)
	return ctx.ES.UpdateFields(elastic.DocContracts, contract.ID, contract, "ProjectID")
}
