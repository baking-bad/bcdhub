package transfer

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
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
	query := buildGetContext(ctx)
	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTransfers}, query, &response); err != nil {
		return po, err
	}

	hits := response.Hits.Hits
	transfers := make([]transfer.Transfer, len(hits))
	for i := range hits {
		if err := json.Unmarshal(hits[i].Source, &transfers[i]); err != nil {
			return po, err
		}
		transfers[i].ID = hits[i].ID
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
	if err = storage.es.Query([]string{models.DocTransfers}, query, &response); err != nil {
		return
	}

	result.Supply = response.Aggs.Result.Value.Supply
	result.Transfered = response.Aggs.Result.Value.Transfered
	return
}

// GetToken24HoursVolume - returns token volume for last 24 hours
func (storage *Storage) GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID int64) (float64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("contract.keyword", contract),
				core.Term("network", network),
				core.Term("status", consts.Applied),
				core.Term("token_id", tokenID),
				core.Range("timestamp", core.Item{
					"lte": "now",
					"gt":  "now-24h",
				}),
				core.In("parent.keyword", entrypoints),
				core.In("initiator.keyword", initiators),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{Name: "volume", Body: core.Sum("amount")},
		),
	).Zero()

	var response aggVolumeSumResponse
	if err := storage.es.Query([]string{models.DocTransfers}, query, &response); err != nil {
		return 0, err
	}

	return response.Aggs.Result.Value, nil
}

// GetTokenVolumeSeries -
func (storage *Storage) GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint) ([][]float64, error) {
	hist := core.Item{
		"date_histogram": core.Item{
			"field":             "timestamp",
			"calendar_interval": period,
		},
	}

	hist.Append("aggs", core.Item{
		"result": core.Item{
			"sum": core.Item{
				"field": "amount",
			},
		},
	})

	matches := []core.Item{
		{
			"script": core.Item{
				"script": core.Item{
					"source": "doc['from.keyword'].value !=  doc['to.keyword'].value",
				},
			},
		},
		core.Match("network", network),
		core.Match("status", "applied"),
		core.Term("token_id", tokenID),
	}
	if len(contracts) > 0 {
		addresses := make([]core.Item, len(contracts))
		for i := range contracts {
			addresses[i] = core.MatchPhrase("contract", contracts[i])
		}
		matches = append(matches, core.Bool(
			core.Should(addresses...),
			core.MinimumShouldMatch(1),
		))
	}

	if len(entrypoints) > 0 {
		addresses := make([]core.Item, 0)
		for i := range entrypoints {
			for j := range entrypoints[i].DexVolumeEntrypoints {
				addresses = append(addresses, core.Bool(
					core.Filter(
						core.MatchPhrase("initiator", entrypoints[i].Address),
						core.Match("parent", entrypoints[i].DexVolumeEntrypoints[j]),
					),
				))
			}
		}
		matches = append(matches, core.Bool(
			core.Should(addresses...),
			core.MinimumShouldMatch(1),
		))
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				matches...,
			),
		),
	).Add(
		core.Aggs(core.AggItem{Name: "hist", Body: hist}),
	).Zero()

	var response getTokenVolumeSeriesResponse
	if err := storage.es.Query([]string{models.DocTransfers}, query, &response); err != nil {
		return nil, err
	}

	histogram := make([][]float64, len(response.Agg.Hist.Buckets))
	for i := range response.Agg.Hist.Buckets {
		item := []float64{
			float64(response.Agg.Hist.Buckets[i].Key),
			response.Agg.Hist.Buckets[i].Result.Value,
		}
		histogram[i] = item
	}
	return histogram, nil
}
