package protocol

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
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

// Get - returns current protocol for `network` and `level` (`hash` is optional, leave empty string for default)
func (storage *Storage) Get(network, hash string, level int64) (p protocol.Protocol, err error) {
	query := storage.DB.Table(models.DocProtocol).Where("network = ?", network)

	if level > -1 {
		query = query.Where("start_level <= ?", level)
	}
	if hash != "" {
		query = query.Where("hash = ?", hash)
	}

	err = query.Order("start_level DESC").First(&p).Error
	return
}

// GetByNetworkWithSort -
func (storage *Storage) GetByNetworkWithSort(network, sortField, order string) (response []protocol.Protocol, err error) {
	orderValue := fmt.Sprintf("%s %s", sortField, order)
	err = storage.DB.Table(models.DocProtocol).Where("network = ?", network).Order(orderValue).Find(&response).Error
	return
}

// GetAll - returns all protocol`s entities
func (storage *Storage) GetAll() (response []protocol.Protocol, err error) {
	err = storage.DB.Table(models.DocProtocol).Find(&response).Error
	return
}
