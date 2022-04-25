package transfer

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(es *core.Postgres) *Storage {
	return &Storage{es}
}

// GetAll -
func (storage *Storage) GetAll(level int64) ([]transfer.Transfer, error) {
	var transfers []transfer.Transfer
	err := storage.DB.Model(&transfers).
		Where("level = ?", level).
		Select(&transfers)
	return transfers, err
}

// GetTransfered -
func (storage *Storage) GetTransfered(contract string, tokenID uint64) (result float64, err error) {
	query := storage.DB.Model().Model((*transfer.Transfer)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("to_id is not null").
		Where("from_id is not null").
		Where("contract = ?", contract).
		Where("token_id = ?", tokenID)

	core.IsApplied(query)
	err = query.Select(&result)
	return
}

// GetToken24HoursVolume - returns token volume for last 24 hours
func (storage *Storage) GetToken24HoursVolume(contract string, initiators, entrypoints []string, tokenID uint64) (float64, error) {
	aDayAgo := time.Now().UTC().AddDate(0, 0, -1)

	var volume float64
	query := storage.DB.Model((*transfer.Transfer)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("timestamp > ?", aDayAgo).
		Where("contract = ?", contract).
		Where("token_id = ?", tokenID)

	if len(entrypoints) > 0 {
		query.WhereIn("parent IN (?)", entrypoints)
	}
	if len(initiators) > 0 {
		var ids []int64
		if err := storage.DB.Model((*account.Account)(nil)).
			Column("id").
			WhereIn("address IN (?)", initiators).
			Select(&ids); err != nil {
			return 0, err
		}
		query.WhereIn("initiator_id IN (?)", ids)
	}

	core.IsApplied(query)
	err := query.Select(&volume)

	return volume, err
}

const (
	tokenVolumeSeriesRequestTemplate = `
		with f as (
			select generate_series(
			date_trunc(?period, ?start_date),
			date_trunc(?period, now()),
			?interval ::interval
			) as val
		)
		select
			extract(epoch from f.val) as date_part,
			sum(amount) as value
		from f
		left join transfers on date_trunc(?period, transfers.timestamp) = f.val where (transfers.from_id != transfers.to_id) and (status = 1) and token_id = ?token_id ?conditions
		group by 1
		order by date_part
	`
)

// GetTokenVolumeSeries -
func (storage *Storage) GetTokenVolumeSeries(period string, contracts []string, entrypoints []dapp.DAppContract, tokenID uint64) ([][]float64, error) {
	if err := core.ValidateHistogramPeriod(period); err != nil {
		return nil, err
	}

	conditions := make([]string, 0)

	if len(contracts) > 0 {
		contractConditions := make([]string, len(contracts))
		for i := range contracts {
			contractConditions[i] = fmt.Sprintf("contract = '%s'", contracts[i])
		}
		conditions = append(conditions, strings.Join(contractConditions, " or "))
	}

	if len(entrypoints) > 0 {
		entrypointConditions := make([]string, 0)
		for _, e := range entrypoints {
			var initiatorID int64
			if err := storage.DB.Model((*account.Account)(nil)).Column("id").Where("address = ?", e.Address).Select(&initiatorID); err != nil {
				return nil, err
			}
			for j := range e.Entrypoint {
				entrypointConditions = append(entrypointConditions, fmt.Sprintf("(initiator_id = %d and parent = '%s')", initiatorID, e.Entrypoint[j]))
			}
		}
		conditions = append(conditions, strings.Join(entrypointConditions, " or "))
	}

	stringConditions := strings.Join(conditions, ") and (")
	if len(stringConditions) > 0 {
		stringConditions = "and (" + stringConditions
		stringConditions += ")"
	}

	var resp []core.HistogramResponse
	if _, err := storage.DB.
		WithParam("token_id", tokenID).
		WithParam("period", period).
		WithParam("start_date", pg.Safe(core.GetHistogramInterval(period))).
		WithParam("interval", fmt.Sprintf("1 %s", period)).
		WithParam("conditions", pg.Safe(stringConditions)).
		Query(&resp, tokenVolumeSeriesRequestTemplate); err != nil {
		return nil, err
	}

	histogram := make([][]float64, 0, len(resp))
	for i := range resp {
		histogram = append(histogram, []float64{resp[i].DatePart * 1000, resp[i].Value})
	}
	return histogram, nil
}

const (
	calcBalanceRequest = `
	select (coalesce(value_to, 0) - coalesce(value_from, 0)) as balance, coalesce(t1.address, t2.address) as address, coalesce(t1.token_id, t2.token_id) as token_id from 
		(select sum(amount) as value_from, accounts.address as address, token_id from transfers left join accounts on from_id = accounts.id where "from_id" is not null and contract = ?contract group by accounts.address, token_id) t1
	full outer join 
		(select sum(amount) as value_to, accounts.address as address, token_id from transfers left join accounts on to_id = accounts.id where "to_id" is not null and contract = ?contract group by accounts.address, token_id) t2
		on t1.address = t2.address and t1.token_id = t2.token_id;`
)

// CalcBalances -
func (storage *Storage) CalcBalances(contract string) ([]transfer.Balance, error) {
	var balances []transfer.Balance
	_, err := storage.DB.
		WithParam("contract", contract).
		Query(&balances, calcBalanceRequest)
	return balances, err
}
