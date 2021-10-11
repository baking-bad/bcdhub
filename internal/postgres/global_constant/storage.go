package global_constant

import (
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
func (storage *Storage) Get(network types.Network, address string) (response global_constant.GlobalConstant, err error) {
	err = storage.DB.Scopes(core.NetworkAndAddress(network, address)).First(&response).Error
	return
}

// All -
func (storage *Storage) All(network types.Network, addresses ...string) (response []global_constant.GlobalConstant, err error) {
	if len(addresses) == 0 {
		return
	}

	query := storage.DB.Scopes(core.Network(network))

	subQuery := storage.DB.Where("address = ?", addresses[0])
	for i := 1; i < len(addresses); i++ {
		subQuery.Or("address = ?", addresses[i])
	}

	err = query.Where(subQuery).Find(&response).Error
	return
}
