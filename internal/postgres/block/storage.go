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
func (storage *Storage) Get(level int64) (block block.Block, err error) {
	err = storage.DB.Model(&block).
		Where("level = ?", level).
		Limit(1).
		Relation("Protocol").
		Select()
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last() (block block.Block, err error) {
	err = storage.DB.Model(&block).
		Order("id desc").
		Limit(1).
		Relation("Protocol").
		Select()
	if storage.IsRecordNotFound(err) {
		err = nil
	}
	return
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	var network types.Network
	err := storage.DB.Model((*block.Block)(nil)).
		Column("block.network").
		Where("block.chain_id = ?", chainID).
		Limit(1).
		Select(&network)

	return network.String(), err
}
