package main

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func recalculateAll(data amqp.Delivery) error {
	contractID := parseID(data.Body)

	c := models.Contract{ID: contractID}
	if err := ctx.ES.GetByID(&c); err != nil {
		if strings.Contains(err.Error(), "[404 Not Found]") {
			return nil
		}
		return errors.Errorf("[recalculateAll] Find contract error: %s", err)
	}
	if err := recalc(c); err != nil {
		return errors.Errorf("[recalculateAll] Compute error message: %s", err)
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
			return errors.Errorf("[recalc] Error during set contract projectID: %s", err)
		}
	}

	if err := h.UpdateContractStats(&contract); err != nil {
		return errors.Errorf("[recalc] Compute contract stats error message: %s", err)
	}

	if _, err := ctx.ES.UpdateDoc(&contract); err != nil {
		return err
	}

	return nil
}
