package services

import (
	"context"
	"sync"

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
func (ch *ContractsHandler) Handle(ctx context.Context, items []models.Model, wg *sync.WaitGroup) error {
	if len(items) == 0 {
		return nil
	}

	logger.Info().Msgf("%3d contracts are processed", len(items))

	return saveSearchModels(ch.Context, items)
}

// Chunk -
func (ch *ContractsHandler) Chunk(lastID, size int64) ([]models.Model, error) {
	contracts, err := getContracts(ch.StorageDB.DB, lastID, size)
	if err != nil {
		return nil, err
	}

	data := make([]models.Model, len(contracts))
	for i := range contracts {
		data[i] = &contracts[i]
	}
	return data, nil
}
