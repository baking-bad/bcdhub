package services

import (
	"context"
	"sync"

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
func (oh *OperationsHandler) Handle(ctx context.Context, items []models.Model, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Str("network", oh.Network.String()).Msgf("%3d operations are processed", len(items))

	return saveSearchModels(ctx, oh.Context, items)
}

// Chunk -
func (oh *OperationsHandler) Chunk(lastID int64, size int) ([]models.Model, error) {
	operations, err := getOperations(oh.StorageDB.DB, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(operations))
	for i := range operations {
		data[i] = &operations[i]
	}
	return data, nil
}
