package transfer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(es *core.Postgres) *Storage {
	return &Storage{es}
}

// Get -
func (storage *Storage) Get(ctx transfer.GetContext) (po transfer.Pageable, err error) {
	po.Transfers = make([]transfer.Transfer, 0)
	query := storage.DB.Table(models.DocTransfers)
	buildGetContext(storage.DB, query, ctx, true)

	if err = query.Find(&po.Transfers).Error; err != nil {
		return
	}
	countQuery := storage.DB.Table(models.DocTransfers)
	buildGetContext(storage.DB, countQuery, ctx, false)
	if err = query.Count(&po.Total).Error; err != nil {
		return
	}

	if len(po.Transfers) > 0 {
		po.LastID = fmt.Sprintf("%d", po.Transfers[len(po.Transfers)-1].IndexedTime)
	}
	return po, nil
}

// GetAll -
func (storage *Storage) GetAll(network string, level int64) ([]transfer.Transfer, error) {
	var transfers []transfer.Transfer
	err := storage.DB.
		Where("network = ?", network).
		Where("level > ?", level).
		Find(&transfers).Error
	return transfers, err
}

// GetTokenSupply -
func (storage *Storage) GetTokenSupply(network, contract string, tokenID uint64) (result transfer.TokenSupply, err error) {
	var supplied, burned float64
	query := storage.DB.Table(models.DocTransfers).
		Scopes(
			core.Token(network, contract, tokenID),
			core.IsApplied,
		).Select("COALESCE(SUM(amount), 0)")

	if err = query.Where("transfers.from = ''").Scan(&supplied).Error; err != nil {
		return
	}
	if err = query.Where("transfers.to = ''").Scan(&burned).Error; err != nil {
		return
	}
	if err = query.Where("transfers.to != '' AND transfers.from != ''").Scan(&result.Transfered).Error; err != nil {
		return
	}

	result.Supply = supplied - burned

	return
}

// GetToken24HoursVolume - returns token volume for last 24 hours
func (storage *Storage) GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID uint64) (float64, error) {
	aDayAgo := time.Now().UTC().AddDate(0, 0, -1).Unix()

	var volume float64
	err := storage.DB.Table(models.DocTransfers).
		Table("SUM(amount) AS volume").
		Scopes(core.Token(network, contract, tokenID)).
		Where("status = ?", consts.Applied).
		Where("parent IN ?", entrypoints).
		Where("initiator IN ?", initiators).
		Where("timestamp > ?", aDayAgo).
		Scan(&volume).Error

	return volume, err
}

// GetTokenVolumeSeries -
// TODO: realize GetTokenVolumeSeries
func (storage *Storage) GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint64) ([][]float64, error) {
	// hist := core.Item{
	// 	"date_histogram": core.Item{
	// 		"field":             "timestamp",
	// 		"calendar_interval": period,
	// 	},
	// }

	// hist.Append("aggs", core.Item{
	// 	"result": core.Item{
	// 		"sum": core.Item{
	// 			"field": "amount",
	// 		},
	// 	},
	// })

	// matches := []core.Item{
	// 	{
	// 		"script": core.Item{
	// 			"script": core.Item{
	// 				"source": "doc['from.keyword'].value !=  doc['to.keyword'].value",
	// 			},
	// 		},
	// 	},
	// 	core.Match("network", network),
	// 	core.Match("status", consts.Applied),
	// 	core.Term("token_id", tokenID),
	// }
	// if len(contracts) > 0 {
	// 	addresses := make([]core.Item, len(contracts))
	// 	for i := range contracts {
	// 		addresses[i] = core.MatchPhrase("contract", contracts[i])
	// 	}
	// 	matches = append(matches, core.Bool(
	// 		core.Should(addresses...),
	// 		core.MinimumShouldMatch(1),
	// 	))
	// }

	// if len(entrypoints) > 0 {
	// 	addresses := make([]core.Item, 0)
	// 	for i := range entrypoints {
	// 		for j := range entrypoints[i].DexVolumeEntrypoints {
	// 			addresses = append(addresses, core.Bool(
	// 				core.Filter(
	// 					core.MatchPhrase("initiator", entrypoints[i].Address),
	// 					core.Match("parent", entrypoints[i].DexVolumeEntrypoints[j]),
	// 				),
	// 			))
	// 		}
	// 	}
	// 	matches = append(matches, core.Bool(
	// 		core.Should(addresses...),
	// 		core.MinimumShouldMatch(1),
	// 	))
	// }

	// query := core.NewQuery().Query(
	// 	core.Bool(
	// 		core.Filter(
	// 			matches...,
	// 		),
	// 	),
	// ).Add(
	// 	core.Aggs(core.AggItem{Name: "hist", Body: hist}),
	// ).Zero()

	// var response getTokenVolumeSeriesResponse
	// if err := storage.es.Query([]string{models.DocTransfers}, query, &response); err != nil {
	// 	return nil, err
	// }

	// histogram := make([][]float64, len(response.Agg.Hist.Buckets))
	// for i := range response.Agg.Hist.Buckets {
	// 	item := []float64{
	// 		float64(response.Agg.Hist.Buckets[i].Key),
	// 		response.Agg.Hist.Buckets[i].Result.Value,
	// 	}
	// 	histogram[i] = item
	// }
	// return histogram, nil
	return nil, nil
}
