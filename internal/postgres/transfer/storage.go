package transfer

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
	storage.buildGetContext(query, ctx, true)

	if err = query.Find(&po.Transfers).Error; err != nil {
		return
	}

	received := len(po.Transfers)
	size := storage.GetPageSize(ctx.Size)
	if ctx.Offset == 0 && size > received {
		po.Total = int64(len(po.Transfers))
	} else {
		countQuery := storage.DB.Table(models.DocTransfers)
		storage.buildGetContext(countQuery, ctx, false)
		if err = countQuery.Count(&po.Total).Error; err != nil {
			return
		}
	}

	if received > 0 {
		po.LastID = fmt.Sprintf("%d", po.Transfers[received-1].ID)
	}
	return po, nil
}

// GetAll -
func (storage *Storage) GetAll(network types.Network, level int64) ([]transfer.Transfer, error) {
	var transfers []transfer.Transfer
	err := storage.DB.Table(models.DocTransfers).
		Where("network = ?", network).
		Where("level > ?", level).
		Find(&transfers).Error
	return transfers, err
}

// GetTransfered -
func (storage *Storage) GetTransfered(network types.Network, contract string, tokenID uint64) (result float64, err error) {
	if err = storage.DB.Table(models.DocTransfers).
		Scopes(
			core.Token(network, contract, tokenID),
			core.IsApplied,
		).
		Select("COALESCE(SUM(amount), 0)").
		Where("transfers.to != '' AND transfers.from != ''").
		Scan(&result).Error; err != nil {
		return
	}

	return
}

// GetToken24HoursVolume - returns token volume for last 24 hours
func (storage *Storage) GetToken24HoursVolume(network types.Network, contract string, initiators, entrypoints []string, tokenID uint64) (float64, error) {
	aDayAgo := time.Now().UTC().AddDate(0, 0, -1)

	var volume float64
	err := storage.DB.Table(models.DocTransfers).
		Select("COALESCE(SUM(amount), 0)").
		Scopes(core.Token(network, contract, tokenID), core.IsApplied).
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
		left join transfers on date_trunc('%s', transfers.timestamp) = f.val where (transfers.from != transfers.to) and (status = 'applied') %s
		group by 1
		order by date_part
	`
)

// GetTokenVolumeSeries -
func (storage *Storage) GetTokenVolumeSeries(network types.Network, period string, contracts []string, entrypoints []dapp.DAppContract, tokenID uint64) ([][]float64, error) {
	if err := core.ValidateHistogramPeriod(period); err != nil {
		return nil, err
	}

	conditions := make([]string, 0)
	conditions = append(conditions, fmt.Sprintf("(token_id = %d)", tokenID))
	if network != types.Empty {
		conditions = append(conditions, fmt.Sprintf("(network = %d)", network))
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
			for j := range entrypoints[i].Entrypoint {
				addresses = append(addresses, fmt.Sprintf("(initiator = '%s' and parent = '%s')", entrypoints[i].Address, entrypoints[i].Entrypoint[j]))
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
		histogram = append(histogram, []float64{resp[i].DatePart, resp[i].Value})
	}
	return histogram, nil
}

const (
	calcBalanceRequest = `
	select (coalesce(value_to, 0) - coalesce(value_from, 0))::varchar(255) as balance, coalesce(t1.address, t2.address) as address, coalesce(t1.token_id, t2.token_id) as token_id from 
		(select sum(amount) as value_from, "from" as address, token_id from transfers where "from" is not null and contract = '%s' and network = '%s' group by "from", token_id) t1
	full outer join 
		(select sum(amount) as value_to, "to" as address, token_id from transfers where "to" is not null and contract = '%s' and network = '%s' group by "to", token_id) t2
		on t1.address = t2.address and t1.token_id = t2.token_id;`
)

// CalcBalances -
func (storage *Storage) CalcBalances(network types.Network, contract string) ([]transfer.Balance, error) {
	request := fmt.Sprintf(calcBalanceRequest, contract, network, contract, network)
	var balances []transfer.Balance
	err := storage.DB.Raw(request).Scan(&balances).Error
	return balances, err
}
