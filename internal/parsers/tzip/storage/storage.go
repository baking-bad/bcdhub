package storage

import "github.com/baking-bad/bcdhub/internal/models/tzip"

// Storage -
type Storage interface {
	Get(value string) (*tzip.TZIP, error)
}
