package schema

import (
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(address string) (schema.Schema, error) {
	data := schema.Schema{ID: address}
	err := storage.db.GetByID(&data)
	return data, err
}
