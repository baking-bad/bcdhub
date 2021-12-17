package block

import (
	"github.com/baking-bad/bcdhub/internal/models/block"
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
func (storage *Storage) Get(network types.Network, level int64) (block block.Block, err error) {
	err = storage.DB.Model(&block).
		Where("block.network = ?", network).
		Where("level = ?", level).
		Limit(1).
		Relation("Protocol.id").
		Select()
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last(network types.Network) (block block.Block, err error) {
	err = storage.DB.Model(&block).
		Where("block.network = ?", network).
		Order("id desc").
		Limit(1).
		Relation("Protocol.id").
		Select()
	if storage.IsRecordNotFound(err) {
		err = nil
		block.Network = network
		return
	}
	return
}

// LastByNetworks - return last block for all networks
func (storage *Storage) LastByNetworks() (response []block.Block, err error) {
	subQuery := storage.DB.Model((*block.Block)(nil)).ColumnExpr("MAX(block.id) as id").Group("network")
	err = storage.DB.Model((*block.Block)(nil)).
		Relation("Protocol.id").
		Where("block.id IN (?)", subQuery).
		Select(&response)
	return
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	var network types.Network
	err := storage.DB.Model((*block.Block)(nil)).
		Column("block.network").
		Where("block.chain_id = ?", chainID).
		Limit(1).
		Relation("Protocol.id").
		Select(&network)

	return network.String(), err
}
