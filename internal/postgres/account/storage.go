package account

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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
func (storage *Storage) Get(address string) (account account.Account, err error) {
	err = storage.DB.Model(&account).
		Where("address = ?", address).
		Limit(1).
		Select(&account)
	return
}

// Alias -
func (storage *Storage) Alias(address string) (alias string, err error) {
	err = storage.DB.Model((*account.Account)(nil)).
		Column("alias").
		Where("address = ?", address).
		Limit(1).
		Select(&alias)
	return
}

// UpdateAlias -
func (storage *Storage) UpdateAlias(account account.Account) error {
	_, err := storage.DB.Model(&account).Set("alias = _data.alias").WherePK().Update()
	return err
}
