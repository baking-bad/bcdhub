package bulk

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}
