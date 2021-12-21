package services

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
)

// ContractsHandler -
type ContractsHandler struct {
	*config.Context
}

// NewContractsHandler -
func NewContractsHandler(ctx *config.Context) *ContractsHandler {
	return &ContractsHandler{ctx}
}

// Handle -
func (ch *ContractsHandler) Handle(ctx context.Context, items []models.Model) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Msgf("%3d contracts are processed", len(items))

	return saveSearchModels(ch.Context, items)
}

// Chunk -
func (ch *ContractsHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	operations, err := getContracts(ch.StorageDB.DB, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(operations))
	for i := range operations {
		data[i] = &operations[i]
	}
	return data, nil
}
