package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// GetTokenMetadata -
func (e *Elastic) GetTokenMetadata(address string, network string, tokenID int64) (token models.TokenMetadata, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("contract", address),
				term("token_id", tokenID),
			),
		),
	).One()

	response, err := e.query([]string{DocTokenMetadata}, query)
	if err != nil {
		return
	}
	if response.Get("hits.total.value").Int() == 0 {
		return token, errors.Errorf("%s: token metadata for %s %s with token ID %d", RecordNotFound, network, address, tokenID)
	}
	hit := response.Get("hits.hits.0")
	token.ParseElasticJSON(hit)
	return
}

// GetTokenMetadatas -
func (e *Elastic) GetTokenMetadatas(address string, network string) ([]models.TokenMetadata, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("contract", address),
			),
		),
	).All()

	response, err := e.query([]string{DocTokenMetadata}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, errors.Errorf("%s: token metadata for %s %s", RecordNotFound, network, address)
	}

	tokens := make([]models.TokenMetadata, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var token models.TokenMetadata
		token.ParseElasticJSON(hit)
		tokens = append(tokens, token)
	}
	return tokens, nil
}

// GetAffectedTokenMetadata -
func (e *Elastic) GetAffectedTokenMetadata(network string, level int64) ([]models.TokenMetadata, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("level", qItem{
					"gt": level,
				}),
			),
		),
	).All()

	response, err := e.query([]string{DocTokenMetadata}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, errors.Errorf("%s: token metadata for %s %d", RecordNotFound, network, level)
	}

	tokens := make([]models.TokenMetadata, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var token models.TokenMetadata
		token.ParseElasticJSON(hit)
		tokens = append(tokens, token)
	}
	return tokens, nil
}
