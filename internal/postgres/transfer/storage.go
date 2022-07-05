package transfer

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
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
