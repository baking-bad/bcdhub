package storage

import "github.com/baking-bad/bcdhub/internal/models"

// Storage -
type Storage interface {
	Get(value string) (*models.TZIP, error)
}
