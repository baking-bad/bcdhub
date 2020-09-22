package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/streadway/amqp"
)

func getContract(data amqp.Delivery) error {
	contractID := parseID(data.Body)

	c := models.Contract{ID: contractID}
	if err := ctx.ES.GetByID(&c); err != nil {
		return errors.Errorf("[getContract] Find contract error: %s", err)
	}

	if err := parseContract(c); err != nil {
		return errors.Errorf("[getContract] Compute error message: %s", err)
	}
	return nil
}

func parseContract(contract models.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if contract.Alias == "" {
		h.SetContractAlias(ctx.Aliases, &contract)
	}

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return errors.Errorf("[parseContract] Error during set contract projectID: %s", err)
		}
	}

	if !contract.Verified {
		if err := h.SetContractVerification(&contract); err != nil {
			return errors.Errorf("[parseContract] Error during set contract verification: %s", err)
		}
	}

	rpc, err := ctx.GetRPC(contract.Network)
	if err != nil {
		return err
	}

	if err := h.CreateTokenMetadata(rpc, ctx.SharePath, &contract); err != nil {
		return err
	}

	logger.Info("Contract %s to project %s", contract.Address, contract.ProjectID)

	return ctx.ES.UpdateFields(elastic.DocContracts, contract.ID, contract, "ProjectID", "Alias", "Verified", "VerificationSource")
}
