package bigmap

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
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
func (storage *Storage) Get(network types.Network, ptr int64, contract string) (*bigmap.BigMap, error) {
	query := storage.DB.Where("network = ?", network).Where("ptr = ?", ptr)
	if contract != "" {
		query.Where("contract = ?", contract)
	}
	b := new(bigmap.BigMap)
	return b, query.First(b).Error
}

// Get -
func (storage *Storage) GetByContract(network types.Network, contract string) (res []bigmap.BigMap, err error) {
	err = storage.DB.Where("network = ?", network).Where("contract = ?", contract).Find(&res).Error
	return
}
