package block

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(network string, level int64) (block block.Block, err error) {
	query := storage.db.Query(models.DocBlocks).
		WhereString("network", reindexer.EQ, network).
		WhereInt64("level", reindexer.EQ, level)

	err = storage.db.GetOne(query, &block)
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last(network string) (block block.Block, err error) {
	query := storage.db.Query(models.DocBlocks).
		WhereString("network", reindexer.EQ, network).
		Sort("level", true)

	err = storage.db.GetOne(query, &block)
	return
}

// LastByNetworks - return last block for all networks
func (storage *Storage) LastByNetworks() ([]block.Block, error) {
	network, err := storage.db.GetUnique("network", storage.db.Query(models.DocBlocks))
	if err != nil {
		return nil, err
	}

	response := make([]block.Block, 0)
	for i := range network {
		blockQuery := storage.db.Query(models.DocBlocks).
			Match("network", network[i]).
			Sort("level", true).
			Limit(1)

		var b block.Block
		if err := storage.db.GetOne(blockQuery, &b); err != nil {
			return nil, err
		}
		response = append(response, b)
	}
	return response, nil
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	query := storage.db.Query(models.DocBlocks).
		WhereString("chain_id", reindexer.EQ, chainID)

	var block block.Block
	err := storage.db.GetOne(query, &block)
	return block.Network, err
}
