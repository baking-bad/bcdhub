package services

import (
	"context"
	"sync"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Service -
type Service interface {
	Init() error
	Start(ctx context.Context)
	Close() error
}

// Handler -
type Handler interface {
	Chunk(lastID int64, size int) ([]models.Model, error)
	Handle(ctx context.Context, items []models.Model, wg *sync.WaitGroup) error
}
