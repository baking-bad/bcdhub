package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/elastic/consts"
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
func (storage *Storage) Get(ctx []tokenmetadata.GetContext, size, offset int64) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := buildGetTokenMetadataContext(ctx...)
	scrollCtx := core.NewScrollContext(storage.es, query, size, consts.DefaultScrollSize)
	scrollCtx.Offset = offset
	err = scrollCtx.Get(&tokens)
	return
}

// Get -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := buildGetTokenMetadataContext(ctx...)
	err = storage.es.GetAllByQuery(query, &tokens)
	return
}
