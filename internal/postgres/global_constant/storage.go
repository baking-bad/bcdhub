package global_constant

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
func (storage *Storage) Get(network types.Network, address string) (response global_constant.GlobalConstant, err error) {
	query := storage.DB.Model(&response)
	core.NetworkAndAddress(network, address)(query)
	err = query.First()
	return
}

// All -
func (storage *Storage) All(network types.Network, addresses ...string) (response []global_constant.GlobalConstant, err error) {
	if len(addresses) == 0 {
		return
	}

	query := storage.DB.Model().Table(models.DocGlobalConstants).Where("address IN (?)", pg.In(addresses))
	core.Network(network)(query)

	err = query.Select(&response)
	return
}
