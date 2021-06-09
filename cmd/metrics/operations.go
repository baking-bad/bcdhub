package main

import (
	"github.com/baking-bad/bcdhub/internal/models"
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
		updated = append(updated, &operations[i])
	}
	logger.Info("%d operations are processed", len(operations))

	return saveSearchModels(ctx.Searcher, updated)
}
