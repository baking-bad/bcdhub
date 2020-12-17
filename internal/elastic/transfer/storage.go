package transfer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

const (
	maxTransfersSize = 10000
)

// Get -
func (storage *Storage) Get(ctx transfer.GetContext) (po transfer.Pageable, err error) {
	query := ctx.Build()
	var response core.SearchResponse
	if err := storage.es.Query([]string{consts.DocTransfers}, query.(core.Base), &response); err != nil {
		return po, err
	}

	hits := response.Hits.Hits
	transfers := make([]transfer.Transfer, len(hits))
	for i := range hits {
		if err := json.Unmarshal(hits[i].Source, &transfers[i]); err != nil {
			return po, err
		}
	}
	po.Transfers = transfers
	po.Total = response.Hits.Total.Value
	if len(transfers) > 0 {
		po.LastID = fmt.Sprintf("%d", transfers[len(transfers)-1].IndexedTime)
	}
	return po, nil
}

// GetAll -
func (storage *Storage) GetAll(network string, level int64) ([]transfer.Transfer, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Range("level", core.Item{"gt": level}),
			),
		),
	)

	transfers := make([]transfer.Transfer, 0)
	err := storage.es.GetAllByQuery(query, &transfers)
	return transfers, err
}

// GetTokenSupply -
func (storage *Storage) GetTokenSupply(network, address string, tokenID int64) (result transfer.TokenSupply, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("contract", address),
				core.Term("token_id", tokenID),
				core.Match("status", "applied"),
			),
		),
	).Add(
		core.Item{
			"aggs": core.Item{
				"result": core.Item{
					"scripted_metric": core.Item{
						"init_script": `state.result = ["supply":0, "transfered":0]`,
						"map_script": `
							if (doc['from.keyword'].value == "") {
								state.result["supply"] = state.result["supply"] + doc["amount"].value;
							} else if (doc['to.keyword'].value == "") {
								state.result["supply"] = state.result["supply"] - doc["amount"].value;
							} else {							
								state.result["transfered"] = state.result["transfered"] + doc["amount"].value;
						}`,
						"combine_script": `return state.result`,
						"reduce_script": `
							Map result = ["supply":0, "transfered":0]; 
							for (state in states) { 
								result["transfered"] = result["transfered"] + state["transfered"];
								result["supply"] = result["supply"] + state["supply"];
							} 
							return result;
						`,
					},
				},
			},
		},
	).Zero()

	var response getTokenSupplyResponse
	if err = storage.es.Query([]string{consts.DocTransfers}, query, &response); err != nil {
		return
	}

	result.Supply = response.Aggs.Result.Value.Supply
	result.Transfered = response.Aggs.Result.Value.Transfered
	return
}

// GetBalances -
func (storage *Storage) GetBalances(network, contract string, level int64, addresses ...transfer.TokenBalance) (map[transfer.TokenBalance]int64, error) {
	filters := []core.Item{
		core.Match("network", network),
	}

	if contract != "" {
		filters = append(filters, core.MatchPhrase("contract", contract))
	}

	if level > 0 {
		filters = append(filters, core.Range("level", core.Item{
			"lt": level,
		}))
	}

	b := core.Bool(
		core.Filter(filters...),
	)

	if len(addresses) > 0 {
		addressFilters := make([]core.Item, 0)

		for _, a := range addresses {
			addressFilters = append(addressFilters, core.Bool(
				core.Filter(
					core.MatchPhrase("from", a.Address),
					core.Term("token_id", a.TokenID),
				),
			))
		}

		b.Get("bool").Extend(
			core.Should(addressFilters...),
		)
		b.Get("bool").Extend(core.MinimumShouldMatch(1))
	}

	query := core.NewQuery().Query(b).Add(
		core.Item{
			"aggs": core.Item{
				"balances": core.Item{
					"scripted_metric": core.Item{
						"init_script": "state.balances = [:]",
						"map_script": `
						if (!state.balances.containsKey(doc['from.keyword'].value)) {
							state.balances[doc['from.keyword'].value + '_' + doc['token_id'].value] = doc['amount'].value;
						} else {
							state.balances[doc['from.keyword'].value + '_' + doc['token_id'].value] = state.balances[doc['from.keyword'].value + '_' + doc['token_id'].value] - doc['amount'].value;
						}
						
						if (!state.balances.containsKey(doc['to.keyword'].value)) {
							state.balances[doc['to.keyword'].value + '_' + doc['token_id'].value] = doc['amount'].value;
						} else {
							state.balances[doc['to.keyword'].value + '_' + doc['token_id'].value] = state.balances[doc['to.keyword'].value + '_' + doc['token_id'].value] + doc['amount'].value;
						}
						`,
						"combine_script": `
						Map balances = [:]; 
						for (entry in state.balances.entrySet()) { 
							if (!balances.containsKey(entry.getKey())) {
								balances[entry.getKey()] = entry.getValue();
							} else {
								balances[entry.getKey()] = balances[entry.getKey()] + entry.getValue();
							}
						} 
						return balances;
						`,
						"reduce_script": `
						Map balances = [:]; 
						for (state in states) { 
							for (entry in state.entrySet()) {
								if (!balances.containsKey(entry.getKey())) {
									balances[entry.getKey()] = entry.getValue();
								} else {
									balances[entry.getKey()] = balances[entry.getKey()] + entry.getValue();
								}
							}
						} 
						return balances;
						`,
					},
				},
			},
		},
	).Zero()
	var response getAccountBalancesResponse
	if err := storage.es.Query([]string{consts.DocTransfers}, query, &response); err != nil {
		return nil, err
	}

	balances := make(map[transfer.TokenBalance]int64)
	for key, balance := range response.Agg.Balances.Value {
		parts := strings.Split(key, "_")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid addressToken key split size: %d", len(parts))
		}
		tokenID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		balances[transfer.TokenBalance{
			Address: parts[0],
			TokenID: tokenID,
		}] = int64(balance)
	}
	return balances, nil
}

// GetToken24HoursVolume - returns token volume for last 24 hours
func (storage *Storage) GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID int64) (float64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				term("contract.keyword", contract),
				term("network", network),
				term("status", consts.Applied),
				term("token_id", tokenID),
				rangeQ("timestamp", qItem{
					"lte": "now",
					"gt":  "now-24h",
				}),
				in("parent.keyword", entrypoints),
				in("initiator.keyword", initiators),
			),
		),
	).Add(
		aggs(
			aggItem{"volume", sum("amount")},
		),
	).Zero()

	var response aggVolumeSumResponse
	if err := e.query([]string{consts.DocTransfers}, query, &response); err != nil {
		return 0, err
	}

	return response.Aggs.Result.Value, nil
}
