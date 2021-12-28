package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
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

func (storage *Storage) Get(network types.Network, contract string, accountID int64, tokenID uint64) (t tokenbalance.TokenBalance, err error) {
	err = storage.DB.Model(&t).
		Where("token_balance.network = ?", network).
		Where("contract = ?", contract).
		Where("account_id = ?", accountID).
		Where("token_id = ?", tokenID).
		Relation("Account").First()

	if storage.IsRecordNotFound(err) {
		t.Network = network
		t.Contract = contract
		t.AccountID = accountID
		t.Account = account.Account{
			Network: network,
		}
		t.TokenID = tokenID
		t.Balance = decimal.Zero
		err = nil
	}

	return
}

// GetHolders -
func (storage *Storage) GetHolders(network types.Network, contract string, tokenID uint64) ([]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance
	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		Where("token_balance.network = ?", network).
		Where("contract = ?", contract).
		Where("token_id = ?", tokenID).
		Where("balance != 0").
		Relation("Account")

	err := query.Select(&balances)
	return balances, err
}

// Batch -
func (storage *Storage) Batch(network types.Network, accountID []int64) (map[string][]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance

	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		WhereIn("account_id IN (?)", accountID).
		Where("network = ?", network).
		Relation("Account.address")

	if err := query.Select(&balances); err != nil {
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
func (storage *Storage) CountByContract(network types.Network, accountID int64, hideEmpty bool) (map[string]int64, error) {
	var resp []tokensByContract
	query := storage.DB.Model((*tokenbalance.TokenBalance)(nil)).
		ColumnExpr("contract, count(*) as tokens_count").
		Where("network = ?", network).
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
func (storage *Storage) TokenSupply(network types.Network, contract string, tokenID uint64) (supply string, err error) {
	query := storage.DB.Model().Table(models.DocTokenBalances).
		ColumnExpr("coalesce(sum(balance), 0)::text as supply")
	core.Token(network, contract, tokenID)(query)

	err = query.Limit(1).Select(&supply)
	return
}
