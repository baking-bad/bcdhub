package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Update -
func (storage *Storage) Update(updates []*tokenbalance.TokenBalance) error {
	if len(updates) == 0 {
		return nil
	}

	return storage.DB.Transaction(func(tx *gorm.DB) error {
		for i := range updates {
			if err := tx.Table(models.DocTokenBalances).Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "network"},
					{Name: "contract"},
					{Name: "address"},
					{Name: "token_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"balance"}),
			}).Create(&updates[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetHolders -
func (storage *Storage) GetHolders(network, contract string, tokenID uint64) ([]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance
	err := storage.DB.Table(models.DocTokenBalances).
		Scopes(core.Token(network, contract, tokenID)).
		Where("balance != '0").
		Find(&balances).Error
	return balances, err
}

// GetAccountBalances -
func (storage *Storage) GetAccountBalances(network, address string) ([]tokenbalance.TokenBalance, error) {
	var balances []tokenbalance.TokenBalance
	err := storage.DB.Table(models.DocTokenBalances).
		Scopes(core.NetworkAndAddress(network, address)).
		Find(&balances).Error
	return balances, err
}

// NFTHolders -
func (storage *Storage) NFTHolders(network, contract string, tokenID uint64) (tokens []tokenbalance.TokenBalance, err error) {
	err = storage.DB.
		Scopes(core.Token(network, contract, tokenID)).
		Where("balance != '0'").
		Find(&tokens).Error
	return
}
