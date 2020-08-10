package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetTokens -
func (e *Elastic) GetTokens(network, tokenInterface string, lastAction, size int64) ([]models.Contract, int64, error) {
	tags := []string{"fa12", "fa1", "fa2"}
	if tokenInterface == "fa12" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = []string{tokenInterface}
	}

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				in("tags", tags),
			),
		),
	).Sort("last_action", "desc").Size(size)

	if lastAction != 0 {
		query = query.SearchAfter([]interface{}{lastAction * 1000})
	}

	result, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return nil, 0, err
	}

	contracts := make([]models.Contract, 0)
	for _, hit := range result.Get("hits.hits").Array() {
		var contract models.Contract
		contract.ParseElasticJSON(hit)
		contracts = append(contracts, contract)
	}
	return contracts, result.Get("hits.total.value").Int(), nil
}

// GetTokenTransferOperations -
func (e *Elastic) GetTokenTransferOperations(network, address, lastID string, size int64) (PageableOperations, error) {
	if size == 0 {
		size = defaultSize
	}
	filterItems := []qItem{
		in("entrypoint", []string{"mint", "transfer"}),
		matchQ("parameter_strings", address),
		matchQ("network", network),
	}
	if lastID != "" {
		filterItems = append(filterItems, rangeQ("indexed_time", qItem{"lt": lastID}))
	}

	query := newQuery().Query(
		boolQ(
			filter(
				filterItems...,
			),
		),
	).Sort("timestamp", "desc").Size(size)

	po := PageableOperations{}
	result, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return po, err
	}

	hits := result.Get("hits.hits").Array()
	operations := make([]models.Operation, len(hits))
	for i, hit := range hits {
		operations[i].ParseElasticJSON(hit)
	}
	po.Operations = operations
	po.LastID = result.Get("hits").Get("hits|@reverse|0").Get("_source.indexed_time").String()
	return po, nil
}

// GetTokensStats -
func (e *Elastic) GetTokensStats(network string, addresses, entrypoints []string) (map[string]TokenUsageStats, error) {
	addressFilters := make([]qItem, len(addresses))
	for i := range addresses {
		addressFilters[i] = matchPhrase("destination", addresses[i])
	}

	entrypointFilters := make([]qItem, len(entrypoints))
	for i := range entrypoints {
		entrypointFilters[i] = matchPhrase("entrypoint", entrypoints[i])
	}

	query := newQuery().Query(
		boolQ(
			must(
				matchQ("network", network),
				boolQ(
					should(addressFilters...),
					minimumShouldMatch(1),
				),
				boolQ(
					should(entrypointFilters...),
					minimumShouldMatch(1),
				),
			),
		),
	).Add(
		aggs("by_dest", qItem{
			"terms": qItem{
				"field": "destination.keyword",
				"size":  maxQuerySize,
			},
			"aggs": qItem{
				"by_entrypoint": qItem{
					"terms": qItem{
						"field": "entrypoint.keyword",
						"size":  maxQuerySize,
					},
					"aggs": qItem{
						"average_consumed_gas": qItem{
							"avg": qItem{"field": "result.consumed_gas"},
						},
					},
				},
			},
		}),
	).Zero()

	result, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return nil, err
	}

	response := make(map[string]TokenUsageStats)
	buckets := result.Get("aggregations.by_dest.buckets").Array()
	for _, bucket := range buckets {
		address := bucket.Get("key").String()
		tokenUsage := make(TokenUsageStats)
		methods := bucket.Get("by_entrypoint.buckets").Array()
		for _, method := range methods {
			key := method.Get("key").String()
			tokenUsage[key] = TokenMethodUsageStats{
				Count:       method.Get("doc_count").Int(),
				ConsumedGas: method.Get("average_consumed_gas.value").Int(),
			}
		}

		response[address] = tokenUsage
	}

	return response, nil
}

// GetTokenVolumeSeries -
func (e *Elastic) GetTokenVolumeSeries(network, period string, address []string, tokenID uint) ([][]int64, error) {
	hist := qItem{
		"date_histogram": qItem{
			"field":             "timestamp",
			"calendar_interval": period,
		},
	}

	hist.Append("aggs", qItem{
		"result": qItem{
			"sum": qItem{
				"field": "amount",
			},
		},
	})

	matches := []qItem{
		matchQ("network", network),
		matchQ("status", "applied"),
	}
	if len(address) > 0 {
		addresses := make([]qItem, len(address))
		for i := range address {
			addresses[i] = matchPhrase("contract", address[i])
		}
		matches = append(matches, boolQ(
			should(addresses...),
			minimumShouldMatch(1),
		))
	}

	query := newQuery().Query(
		boolQ(
			filter(
				matches...,
			),
		),
	).Add(
		aggs("hist", hist),
	).Zero()

	response, err := e.query([]string{DocTransfers}, query)
	if err != nil {
		return nil, err
	}

	data := response.Get("aggregations.hist.buckets").Array()
	histogram := make([][]int64, 0)
	for _, hit := range data {
		item := []int64{
			hit.Get("key").Int(),
			hit.Get("result.value").Int(),
		}
		histogram = append(histogram, item)
	}
	return histogram, nil
}
