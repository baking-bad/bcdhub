package services

import (
	"context"

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
	Chunk(lastID, size int64) ([]models.Model, error)
	Handle(ctx context.Context, items []models.Model) error
}
