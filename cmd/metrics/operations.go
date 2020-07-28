package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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

	op := models.Operation{ID: operationID}
	if err := ctx.ES.GetByID(&op); err != nil {
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

	if strings.HasPrefix(operation.Destination, "KT") || operation.Kind == consts.Origination {
		if err := h.SetBigMapDiffsStrings(operation.ID); err != nil {
			return err
		}
	}

	logger.Info("Operation %s processed", operation.ID)
	return nil
}
