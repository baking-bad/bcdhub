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
type Handler[M models.Constraint] interface {
	Chunk(lastID int64, size int) ([]M, error)
	Handle(ctx context.Context, items []M, wg *sync.WaitGroup) error
}
