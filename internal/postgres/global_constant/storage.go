package global_constant

import (
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
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
func (storage *Storage) Get(address string) (response contract.GlobalConstant, err error) {
	query := storage.DB.Model(&response)
	core.Address(address)(query)
	err = query.First()
	return
}

// All -
func (storage *Storage) All(addresses ...string) (response []contract.GlobalConstant, err error) {
	if len(addresses) == 0 {
		return
	}

	err = storage.DB.Model((*contract.GlobalConstant)(nil)).Where("address IN (?)", pg.In(addresses)).Select(&response)
	return
}

// List -
func (storage *Storage) List(size, offset int64) (response []contract.GlobalConstant, err error) {
	if offset < 0 {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	err = storage.DB.Model((*contract.GlobalConstant)(nil)).
		Limit(int(size)).
		Offset(int(offset)).
		Order("id desc").
		Select(&response)
	return

}

// ForContract -
func (storage *Storage) ForContract(address string, size, offset int64) (response []contract.GlobalConstant, err error) {
	if offset < 0 || address == "" {
		return nil, nil
	}
	if size < 1 || size > consts.MaxSize {
		size = consts.DefaultSize
	}

	err = storage.DB.Model((*contract.Contract)(nil)).
		ColumnExpr("global_constants.*").
		Join("LEFT JOIN accounts on account_id = accounts.id").
		Join("LEFT JOIN script_constants on script_constants.script_id = jakarta_id").
		Join("LEFT JOIN global_constants on script_constants.global_constant_id = global_constants.id").
		Where("accounts.address = ?", address).
		Limit(int(size)).
		Offset(int(offset)).
		Order("id desc").
		Select(&response)
	return
}
