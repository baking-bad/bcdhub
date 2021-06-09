package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models"
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

func (storage *Storage) Get(network types.Network, contract, address string, tokenID uint64) (t tokenbalance.TokenBalance, err error) {
	err = storage.DB.Table(models.DocTokenBalances).
		Scopes(core.Token(network, contract, tokenID)).
		Where("address = ?", address).
		First(&t).Error

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
	err := storage.DB.Table(models.DocTokenBalances).
		Scopes(core.Token(network, contract, tokenID)).
		Where("balance != 0").
		Find(&balances).Error
	return balances, err
}

// Batch -
func (storage *Storage) Batch(network types.Network, addresses []string) (map[string][]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance

	query := storage.DB.Table(models.DocTokenBalances).Scopes(core.Network(network))

	for i := range addresses {
		if i == 0 {
			query.Where("address = ?", addresses[i])
		} else {
			query.Or("address = ?", addresses[i])
		}
	}

	if err := query.Find(&balances).Error; err != nil {
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
func (storage *Storage) CountByContract(network types.Network, address string) (map[string]int64, error) {
	var resp []tokensByContract
	query := storage.DB.Table(models.DocTokenBalances).
		Select("contract, count(*) as tokens_count").
		Scopes(core.NetworkAndAddress(network, address)).
		Group("contract").
		Scan(&resp)

	if query.Error != nil {
		return nil, query.Error
	}

	result := make(map[string]int64)
	for i := range resp {
		result[resp[i].Contract] = resp[i].TokensCount
	}
	return result, nil
}

// TokenSupply -
func (storage *Storage) TokenSupply(network types.Network, contract string, tokenID uint64) (supply string, err error) {
	query := storage.DB.Table(models.DocTokenBalances).
		Select("coalesce(sum(balance), 0)::text as supply").
		Scopes(core.Token(network, contract, tokenID)).
		Scan(&supply)

	err = query.Error
	return
}
