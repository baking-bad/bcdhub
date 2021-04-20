package tokenmetadata

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/models"

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
	query := buildGetTokenMetadataContext(ctx, true)
	size = core.GetSize(size, storage.es.MaxPageSize)
	scrollCtx := core.NewScrollContext(storage.es, query, size, consts.DefaultScrollSize)
	scrollCtx.Offset = offset
	err = scrollCtx.Get(&tokens)
	return
}

// GetAll -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := buildGetTokenMetadataContext(ctx, true)
	err = storage.es.GetAllByQuery(query, &tokens)
	return
}

// GetWithExtras -
func (storage *Storage) GetWithExtras() ([]tokenmetadata.TokenMetadata, error) {
	var tokens []tokenmetadata.TokenMetadata
	if err := storage.es.GetAll(&tokens); err != nil {
		return nil, err
	}

	withExtras := make([]tokenmetadata.TokenMetadata, 0)
	for i := range tokens {
		if len(tokens[i].Extras) > 0 {
			withExtras = append(withExtras, tokens[i])
		}
	}

	return withExtras, nil
}

// Count -
func (storage *Storage) Count(ctx []tokenmetadata.GetContext) (int64, error) {
	return storage.es.CountItems([]string{models.DocTokenMetadata}, buildGetTokenMetadataContext(ctx, false))
}

// GetOne -
func (storage *Storage) GetOne(network, contract string, tokenID int64) (token tokenmetadata.TokenMetadata, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("contract", contract),
				core.Term("token_id", tokenID),
			),
		),
	).One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocTokenMetadata}, query, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		return token, core.NewRecordNotFoundError(models.DocTokenMetadata, "")
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &token)
	return
}
