package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/streadway/amqp"
)

func getOperation(data amqp.Delivery) error {
	var operationID string
	if err := json.Unmarshal(data.Body, &operationID); err != nil {
		return fmt.Errorf("[getOperation] Unmarshal message body error: %s", err)
	}

	op, err := ctx.ES.GetOperationByID(operationID)
	if err != nil {
		return fmt.Errorf("[getOperation] Find operation error: %s", err)
	}

	if err := parseOperation(op); err != nil {
		return fmt.Errorf("[getOperation] Compute error message: %s", err)
	}

	return nil
}

func parseOperation(operation models.Operation) error {
	h := metrics.New(ctx.ES, ctx.DB)

	h.SetOperationAliases(ctx.Aliases, &operation)
	h.SetOperationBurned(&operation)
	h.SetOperationStrings(&operation)

	if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operation.ID, operation); err != nil {
		return err
	}

	for _, address := range []string{operation.Source, operation.Destination} {
		if strings.HasPrefix(address, "KT") {
			if err := setOperationStats(h, address, operation); err != nil {
				return fmt.Errorf("[parseOperation] Compute error message: %s", err)
			}
		}
	}

	if strings.HasPrefix(operation.Destination, "KT") {
		if err := h.SetBigMapDiffsStrings(operation.ID); err != nil {
			return err
		}
	}

	logger.Info("Operation %s processed", operation.ID)
	return nil
}

func setOperationStats(h *metrics.Handler, address string, operation models.Operation) error {
	c, err := ctx.ES.GetContract(map[string]interface{}{
		"network": operation.Network,
		"address": address,
	})

	if err != nil {
		if strings.Contains(err.Error(), "Unknown contract") {
			return nil
		}
		return fmt.Errorf("[setOperationStats] Find contract error: %s", err)
	}

	if err := h.SetContractStats(operation, &c); err != nil {
		return fmt.Errorf("[setOperationStats] compute contract stats error message: %s", err)
	}

	if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, c.ID, c); err != nil {
		return err
	}

	return nil
}
