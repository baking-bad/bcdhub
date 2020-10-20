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
		filters = append(filters, matchPhrase("address", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, matchQ("network", ctx.Network))
	}
	if ctx.Level.Comparator != "" {
		filters = append(filters, rangeQ("level", qItem{
			ctx.Level.Comparator: ctx.Level.Value,
		}))
	}
	if ctx.TokenID != -1 {
		filters = append(filters, term(
			"tokens.token_id", ctx.TokenID,
		))
	}
	return newQuery().Query(
		boolQ(
			filter(filters...),
		),
	).All()
}

// TokenMetadata -
type TokenMetadata struct {
	Address         string
	Network         string
	Symbol          string
	Name            string
	TokenID         int64
	Decimals        int64
	RegistryAddress string
	Extras          map[string]interface{}
}

// GetTokenMetadata -
func (e *Elastic) GetTokenMetadata(ctx GetTokenMetadataContext) (tokens []TokenMetadata, err error) {
	response, err := e.query([]string{DocTZIP}, ctx.buildQuery())
	if err != nil {
		return
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, errors.Errorf("%s: token metadata", RecordNotFound)
	}

	tokens = make([]TokenMetadata, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var tzip models.TZIP
		tzip.ParseElasticJSON(hit)

		for i := range tzip.Tokens.Static {
			tokens = append(tokens, TokenMetadata{
				Address:         tzip.Address,
				Network:         tzip.Network,
				RegistryAddress: tzip.Tokens.Static[i].RegistryAddress,
				Symbol:          tzip.Tokens.Static[i].Symbol,
				Name:            tzip.Tokens.Static[i].Name,
				Decimals:        tzip.Tokens.Static[i].Decimals,
				TokenID:         tzip.Tokens.Static[i].TokenID,
				Extras:          tzip.Tokens.Static[i].Extras,
			})
		}
	}
	return
}

// GetTZIP -
func (e *Elastic) GetTZIP(network, address string) (t models.TZIP, err error) {
	t.Address = address
	t.Network = network
	err = e.GetByID(&t)
	return
}
