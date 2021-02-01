package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
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
func (storage *Storage) Get(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := buildGetTokenMetadataContext(ctx...)
	err = storage.es.GetAllByQuery(query, &tokens)
	return
}
