package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// GetTokenMetadataContext -
type GetTokenMetadataContext struct {
	Contract string
	Network  string
	TokenID  int64
	Level    Range
}

// Range -
type Range struct {
	Comparator string
	Value      int64
}

func (ctx GetTokenMetadataContext) buildQuery() base {
	filters := make([]qItem, 0)

	if ctx.Contract != "" {
		filters = append(filters, matchPhrase("contract", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, matchQ("network", ctx.Network))
	}
	if ctx.TokenID >= 0 {
		filters = append(filters, term("token_id", ctx.TokenID))
	}
	if ctx.Level.Comparator != "" {
		filters = append(filters, rangeQ("level", qItem{
			ctx.Level.Comparator: ctx.Level.Value,
		}))
	}
	return newQuery().Query(
		boolQ(
			filter(filters...),
		),
	).All()
}

// GetTokenMetadata -
func (e *Elastic) GetTokenMetadata(ctx GetTokenMetadataContext) (tokens []models.TokenMetadata, err error) {
	response, err := e.query([]string{DocTokenMetadata}, ctx.buildQuery())
	if err != nil {
		return
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, errors.Errorf("%s: token metadata", RecordNotFound)
	}

	tokens = make([]models.TokenMetadata, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var token models.TokenMetadata
		token.ParseElasticJSON(hit)
		tokens = append(tokens, token)
	}
	return
}
