package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// GetTokenMetadataContext -
type GetTokenMetadataContext struct {
	Contract string
	Network  string
	TokenID  int64
	Level    Range
}

func (ctx GetTokenMetadataContext) buildQuery() base {
	filters := make([]qItem, 0)

	if ctx.Contract != "" {
		filters = append(filters, matchPhrase("address", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, matchQ("network", ctx.Network))
	}
	if ctx.Level.isFilled() {
		filters = append(filters, ctx.Level.build())
	}
	if ctx.TokenID != -1 {
		filters = append(filters, term(
			"tokens.static.token_id", ctx.TokenID,
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
	Level           int64
	Symbol          string
	Name            string
	TokenID         int64
	Decimals        *int64
	RegistryAddress string
	Extras          map[string]interface{}
}

// GetTokenMetadata -
func (e *Elastic) GetTokenMetadata(ctx GetTokenMetadataContext) (tokens []TokenMetadata, err error) {
	tzips := make([]models.TZIP, 0)
	query := ctx.buildQuery()
	if err = e.getAllByQuery(query, &tzips); err != nil {
		return
	}
	if len(tzips) == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	tokens = make([]TokenMetadata, 0)
	for _, tzip := range tzips {
		if tzip.Tokens == nil {
			continue
		}

		for i := range tzip.Tokens.Static {
			tokens = append(tokens, TokenMetadata{
				Address:         tzip.Address,
				Network:         tzip.Network,
				Level:           tzip.Level,
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

// GetDApps -
func (e *Elastic) GetDApps() ([]tzip.DApp, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				exists("dapps"),
			),
		),
	).Sort("dapps.order", "asc").All()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response, "dapps"); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	tokens := make([]tzip.DApp, 0)
	for _, hit := range response.Hits.Hits {
		var model models.TZIP
		if err := json.Unmarshal(hit.Source, &model); err != nil {
			return nil, err
		}
		tokens = append(tokens, model.DApps...)
	}

	return tokens, nil
}

// GetDAppBySlug -
func (e *Elastic) GetDAppBySlug(slug string) (*tzip.DApp, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("dapps.slug", slug),
			),
		),
	).One()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response, "dapps"); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var model models.TZIP
	if err := json.Unmarshal(response.Hits.Hits[0].Source, &model); err != nil {
		return nil, err
	}
	return &model.DApps[0], nil
}

// GetBySlug -
func (e *Elastic) GetBySlug(slug string) (*models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				term("slug.keyword", slug),
			),
		),
	).One()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var data models.TZIP
	err := json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return &data, err
}

// GetAliasesMap -
func (e *Elastic) GetAliasesMap(network string) (map[string]string, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).All()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	aliases := make(map[string]string)
	for _, hit := range response.Hits.Hits {
		var data models.TZIP
		if err := json.Unmarshal(hit.Source, &data); err != nil {
			return nil, err
		}
		aliases[data.Address] = data.Name
	}

	return aliases, nil
}

// GetAliases -
func (e *Elastic) GetAliases(network string) ([]models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				exists("name"),
			),
		),
	).All()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	aliases := make([]models.TZIP, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &aliases[i]); err != nil {
			return nil, err
		}
	}
	return aliases, nil
}

// GetAlias -
func (e *Elastic) GetAlias(network, address string) (*models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
			),
		),
	).One()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var data models.TZIP
	err := json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return &data, err
}

// GetTZIPWithEvents -
func (e *Elastic) GetTZIPWithEvents() ([]models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				exists("events"),
			),
		),
	).All()

	var response SearchResponse
	if err := e.query([]string{DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	tokens := make([]models.TZIP, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &tokens[i]); err != nil {
			return nil, err
		}
	}
	return tokens, nil
}
