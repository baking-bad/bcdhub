package services

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
)

// OperationsHandler -
type OperationsHandler struct {
	*config.Context
}

// NewOperationsHandler -
func NewOperationsHandler(ctx *config.Context) *OperationsHandler {
	return &OperationsHandler{ctx}
}

// Handle -
func (oh *OperationsHandler) Handle(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Msgf("%2d operations are processed", len(items))

	return saveSearchModels(oh.Context, items)
}

// Chunk -
func (oh *OperationsHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	operations, err := getOperations(oh.StorageDB.DB, models.DocOperations, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(operations))
	for i := range operations {
		data[i] = &operations[i]
	}
	return data, nil
}
