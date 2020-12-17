package schema

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/schema"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// Get -
func (storage *Storage) Get(address string) (schema.Schema, error) {
	data := schema.Schema{ID: address}
	err := storage.es.GetByID(&data)
	return data, err
}
