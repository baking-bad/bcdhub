package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/streadway/amqp"
)

func recalculateAll(data amqp.Delivery) error {
	var contractID string
	if err := json.Unmarshal(data.Body, &contractID); err != nil {
		return fmt.Errorf("[recalculateAll] Unmarshal message body error: %s", err)
	}

	c, err := ctx.ES.GetContractByID(contractID)
	if err != nil {
		if strings.Contains(err.Error(), "[404 Not Found]") {
			return nil
		}
		return fmt.Errorf("[recalculateAll] Find contract error: %s", err)
	}
	if err := recalc(c); err != nil {
		return fmt.Errorf("[recalculateAll] Compute error message: %s", err)
	}

	logger.Info("[%s] Contract metrics are recalculated", c.Address)
	return nil
}

func recalc(contract models.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if contract.Alias == "" {
		h.SetContractAlias(ctx.Aliases, &contract)
	}

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(&contract); err != nil {
			return fmt.Errorf("[recalc] Error during set contract projectID: %s", err)
		}
	}

	if err := h.UpdateContractStats(&contract); err != nil {
		return fmt.Errorf("[recalc] Compute contract stats error message: %s", err)
	}

	if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, contract.ID, contract); err != nil {
		return err
	}

	return nil
}
