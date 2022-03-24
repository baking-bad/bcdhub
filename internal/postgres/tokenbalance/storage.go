package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/shopspring/decimal"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(contract string, accountID int64, tokenID uint64) (t tokenbalance.TokenBalance, err error) {
	err = storage.DB.Model(&t).
		Where("contract = ?", contract).
		Where("account_id = ?", accountID).
		Where("token_id = ?", tokenID).
		First()

	if err != nil {
		if storage.IsRecordNotFound(err) {
			t.Contract = contract
			t.AccountID = accountID
			t.Account = account.Account{}
			t.TokenID = tokenID
			t.Balance = decimal.Zero
			err = nil
		}
		return
	}

	err = storage.DB.Model((*account.Account)(nil)).Where("id = ?", t.AccountID).Select(&t.Account)
	return
}

// GetHolders -
func (storage *Storage) GetHolders(contract string, tokenID uint64) ([]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance
	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		Where("contract = ?", contract).
		Where("token_id = ?", tokenID).
		Where("balance != 0")

	err := storage.DB.Model().TableExpr("(?) as token_balance", query).
		ColumnExpr("token_balance.*").
		ColumnExpr("account.id as account__id, account.address as account__address, account.alias as account__alias, account.type as account__type").
		Join("LEFT JOIN accounts as account ON token_balance.account_id = account.id").
		Select(&balances)
	return balances, err
}

// Batch -
func (storage *Storage) Batch(accountID []int64) (map[string][]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance

	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		WhereIn("account_id IN (?)", accountID)

	if err := storage.DB.Model().TableExpr("(?) as token_balance", query).
		ColumnExpr("token_balance.*").
		ColumnExpr("account.address as account__address").
		Join("LEFT JOIN accounts as account ON token_balance.account_id = account.id").
		Select(&balances); err != nil {
		return nil, err
	}

	result := make(map[string][]tokenbalance.TokenBalance)

	for _, b := range balances {
		if _, ok := result[b.Account.Address]; !ok {
			result[b.Account.Address] = make([]tokenbalance.TokenBalance, 0)
		}
		result[b.Account.Address] = append(result[b.Account.Address], b)
	}

	return result, nil
}

type tokensByContract struct {
	Contract    string
	TokensCount int64
}

// CountByContract -
func (storage *Storage) CountByContract(accountID int64, hideEmpty bool) (map[string]int64, error) {
	var resp []tokensByContract
	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		ColumnExpr("contract, count(*) as tokens_count").
		Where("account_id = ?", accountID)

	if hideEmpty {
		query.Where("balance != 0")
	}

	if err := query.Group("contract").Select(&resp); err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for i := range resp {
		result[resp[i].Contract] = resp[i].TokensCount
	}
	return result, nil
}

// TokenSupply -
func (storage *Storage) TokenSupply(contract string, tokenID uint64) (supply string, err error) {
	err = storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		ColumnExpr("coalesce(sum(balance), 0)::text as supply").
		Where("contract = ?", contract).
		Where("token_id = ?", tokenID).
		Select(&supply)
	return
}
