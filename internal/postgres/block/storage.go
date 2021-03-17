package block

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
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
func (storage *Storage) Get(network string, level int64) (block block.Block, err error) {
	err = storage.DB.Table(models.DocBlocks).Scopes(core.Network(network)).Where("level = ?", level).First(&block).Error
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last(network string) (block block.Block, err error) {
	err = storage.DB.Table(models.DocBlocks).Scopes(core.Network(network)).Order("level DESC").First(&block).Error
	if storage.IsRecordNotFound(err) {
		err = nil
		block.Network = network
		return
	}
	return
}

// LastByNetworks - return last block for all networks
func (storage *Storage) LastByNetworks() (response []block.Block, err error) {
	err = storage.DB.Table(models.DocBlocks).Group("network").Find(&response).Error
	return
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	var network string
	err := storage.DB.Table(models.DocBlocks).
		Select("network").
		Where("chain_id = ?", chainID).
		First(&network).Error

	return network, err
}
