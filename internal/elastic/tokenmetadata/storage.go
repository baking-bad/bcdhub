package tokenmetadata

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
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

// GetWithExtras -
func (storage *Storage) GetWithExtras() ([]tokenmetadata.TokenMetadata, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Should(
				core.Exists("extras.description"),
				core.Exists("extras.artifactUri"),
				core.Exists("extras.displayUri"),
				core.Exists("extras.thumbnailUri"),
				core.Exists("extras.externalUri"),
				core.Exists("extras.isTransferable"),
				core.Exists("extras.isBooleanAmount"),
				core.Exists("extras.shouldPreferSymbol"),
			),
			core.MinimumShouldMatch(1),
		),
	).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTokenMetadata}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTokenMetadata, "")
	}

	tokens := make([]tokenmetadata.TokenMetadata, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &tokens[i]); err != nil {
			return nil, err
		}
	}
	return tokens, nil
}
