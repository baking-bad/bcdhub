package main

import (
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func getOperation(ids []int64) error {
	operations, err := ctx.Operations.GetByIDs(ids...)
	if err != nil {
		return errors.Errorf("[getOperation] Find operation error for IDs %v: %s", ids, err)
	}

	updated := make([]models.Model, 0)
	for i := range operations {
		parseOperation(operations[i])
		updated = append(updated, &operations[i])
	}
	logger.Info("%d operations are processed", len(operations))

	if err := saveSearchModels(ctx.Searcher, updated); err != nil {
		return err
	}

	return ctx.Storage.Save(updated)
}

func parseOperation(operation operation.Operation) {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage)
	h.SetOperationStrings(&operation)
}
