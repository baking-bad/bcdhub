package services

import "github.com/baking-bad/bcdhub/internal/models"

// Service -
type Service interface {
	Init() error
	Start()
	Close() error
}

// Handler -
type Handler interface {
	Chunk(lastID, size int64) ([]models.Model, error)
	Handle(items []models.Model) error
}
