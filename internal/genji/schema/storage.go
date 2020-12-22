package schema

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models/schema"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(address string) (schema.Schema, error) {
	data := schema.Schema{ID: address}
	err := storage.db.GetByID(&data)
	return data, err
}
