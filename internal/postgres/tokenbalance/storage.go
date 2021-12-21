package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
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

func (storage *Storage) Get(network types.Network, contract, address string, tokenID uint64) (t tokenbalance.TokenBalance, err error) {
	query := storage.DB.Model(&t).Where("address = ?", address)
	core.Token(network, contract, tokenID)(query)

	err = query.First()

	if storage.IsRecordNotFound(err) {
		t.Network = network
		t.Contract = contract
		t.Address = address
		t.TokenID = tokenID
		t.Balance = decimal.Zero
		err = nil
	}

	return
}

// GetHolders -
func (storage *Storage) GetHolders(network types.Network, contract string, tokenID uint64) ([]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance
	query := storage.DB.Model().Table(models.DocTokenBalances).Where("balance != 0")

	core.Token(network, contract, tokenID)(query)
	err := query.Select(&balances)
	return balances, err
}

// Batch -
func (storage *Storage) Batch(network types.Network, addresses []string) (map[string][]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance

	query := storage.DB.Model().Table(models.DocTokenBalances).Where("address IN (?)", pg.In(addresses))
	core.Network(network)(query)

	if err := query.Select(&balances); err != nil {
		return nil, err
	}

	result := make(map[string][]tokenbalance.TokenBalance)

	for _, b := range balances {
		if _, ok := result[b.Address]; !ok {
			result[b.Address] = make([]tokenbalance.TokenBalance, 0)
		}
		result[b.Address] = append(result[b.Address], b)
	}

	return result, nil
}

type tokensByContract struct {
	Contract    string
	TokensCount int64
}

// CountByContract -
func (storage *Storage) CountByContract(network types.Network, address string, hideEmpty bool) (map[string]int64, error) {
	var resp []tokensByContract
	query := storage.DB.Model().Table(models.DocTokenBalances).
		ColumnExpr("contract, count(*) as tokens_count")
	core.NetworkAndAddress(network, address)(query)

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
