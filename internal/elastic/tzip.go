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
	tzips := make([]models.TZIP, 0)
	query := ctx.buildQuery()
	if err = e.getAllByQuery(query, &tzips); err != nil {
		return
	}
	if len(tzips) > 0 {
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
	).All()

	response, err := e.query([]string{DocTZIP}, query, "dapps")
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	tokens := make([]tzip.DApp, 0)
	for _, hit := range response.Get("hits.hits.#._source.dapps.0").Array() {
		var dapp tzip.DApp
		dapp.ParseElasticJSON(hit)
		tokens = append(tokens, dapp)
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

	response, err := e.query([]string{DocTZIP}, query, "dapps")
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var data tzip.DApp
	data.ParseElasticJSON(response.Get("hits.hits.0._source.dapps.0"))
	return &data, nil
}

// GetBySlug -
func (e *Elastic) GetBySlug(slug string) (*models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchPhrase("slug", slug),
			),
		),
	).One()

	response, err := e.query([]string{DocTZIP}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var data models.TZIP
	data.ParseElasticJSON(response.Get("hits.hits.0"))
	return &data, nil
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

	response, err := e.query([]string{DocTZIP}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	aliases := make(map[string]string)
	for _, hit := range response.Get("hits.hits").Array() {
		var data models.TZIP
		data.ParseElasticJSON(hit)

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

	response, err := e.query([]string{DocTZIP}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	aliases := make([]models.TZIP, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var data models.TZIP
		data.ParseElasticJSON(hit)
		aliases = append(aliases, data)
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

	response, err := e.query([]string{DocTZIP}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	var data models.TZIP
	data.ParseElasticJSON(response.Get("hits.hits.0"))
	return &data, nil
}

// GetTZIPWithViews -
func (e *Elastic) GetTZIPWithViews() ([]models.TZIP, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				exists("events"),
			),
		),
	).All()

	response, err := e.query([]string{DocTZIP}, query)
	if err != nil {
		return nil, err
	}
	if response.Get("hits.total.value").Int() == 0 {
		return nil, NewRecordNotFoundError(DocTZIP, "", query)
	}

	tokens := make([]models.TZIP, 0)
	for _, hit := range response.Get("hits.hits.#._source").Array() {
		var data models.TZIP
		data.ParseElasticJSON(hit)
		tokens = append(tokens, data)
	}

	return tokens, nil
}
