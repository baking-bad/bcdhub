package transfer

import (
	"fmt"
	"strings"
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
	err := storage.DB.Table(models.DocTransfers).
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

const (
	tokenVolumeSeriesRequestTemplate = `
		with f as (
			select generate_series(
			date_trunc('%s', date '2018-06-25'),
			date_trunc('%s', now()),
			'1 %s'::interval
			) as val
		)
		select
			extract(epoch from f.val),
			sum(amount) as value
		from f
		left join transfers on date_trunc('%s', transfers.timestamp) = f.val where (from != to) and (status = 'applied') %s
		group by 1
		order by date_part
	`
)

// GetTokenVolumeSeries -
func (storage *Storage) GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint64) ([][]float64, error) {
	if err := core.ValidateHistogramPeriod(period); err != nil {
		return nil, err
	}

	conditions := make([]string, 0)
	conditions = append(conditions, fmt.Sprintf("(token_id = %d)", tokenID))
	if network != "" {
		conditions = append(conditions, fmt.Sprintf("(network = '%s')", network))
	}

	if len(contracts) > 0 {
		addresses := make([]string, 0)
		for i := range contracts {
			addresses = append(addresses, fmt.Sprintf("(contract = '%s')", contracts[i]))
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(addresses, " or ")))
	}

	if len(entrypoints) > 0 {
		addresses := make([]string, 0)
		for i := range entrypoints {
			for j := range entrypoints[i].DexVolumeEntrypoints {
				addresses = append(addresses, fmt.Sprintf("(initiator = '%s' and parent = '%s')", entrypoints[i].Address, entrypoints[i].DexVolumeEntrypoints[j]))
			}
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(addresses, " or ")))
	}

	var cond string
	if len(conditions) > 0 {
		cond = fmt.Sprintf(" and %s", strings.Join(conditions, " and "))
	}

	req := fmt.Sprintf(tokenVolumeSeriesRequestTemplate, period, period, period, period, cond)

	var resp []core.HistogramResponse
	if err := storage.DB.Raw(req).Scan(&resp).Error; err != nil {
		return nil, err
	}

	histogram := make([][]float64, 0, len(resp))
	for i := range resp {
		histogram = append(histogram, []float64{resp[i].DatePart, float64(resp[i].Value)})
	}
	return histogram, nil
}
