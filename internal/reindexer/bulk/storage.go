package bulk

import (
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
