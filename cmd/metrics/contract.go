package main

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/metrics"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/streadway/amqp"
)

func getContract(data amqp.Delivery) error {
	var contractID string
	if err := json.Unmarshal(data.Body, &contractID); err != nil {
		return fmt.Errorf("[getContract] Unmarshal message body error: %s", err)
	}

	c, err := ctx.ES.GetContractByID(contractID)
	if err != nil {
		return fmt.Errorf("[getContract] Find contract error: %s", err)
	}

	if err := parseContract(c); err != nil {
		return fmt.Errorf("[getContract] Compute error message: %s", err)
	}
	return nil
}

func parseContract(contract models.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if contract.Alias == "" {
		h.SetContractAlias(&contract, ctx.Aliases)
	}

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return fmt.Errorf("[parseContract] Error during set contract projectID: %s", err)
		}
	}

	logger.Info("Contract %s to project %s", contract.Address, contract.ProjectID)

	if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, contract.ID, contract); err != nil {
		return err
	}
	return nil
}
