package block

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// GetBlock -
func (storage *Storage) GetBlock(network string, level int64) (block block.Block, err error) {
	builder := core.NewBuilder().SelectAll(models.DocBlocks).And(
		core.NewEq("network", network),
		core.NewEq("level", level),
	).One()

	err = storage.db.GetOne(builder, &block)
	return
}

// GetLastBlock - returns current indexer state for network
func (storage *Storage) GetLastBlock(network string) (block block.Block, err error) {
	builder := core.NewBuilder().SelectAll(models.DocBlocks).And(
		core.NewEq("network", network),
	).SortDesc("level").One()

	err = storage.db.GetOne(builder, &block)
	return
}

// GetLastBlocks - return last block for all networks
func (storage *Storage) GetLastBlocks() ([]block.Block, error) {
	// builder := core.NewBuilder().SelectAll(models.DocBlocks).GroupBy("network")

	// query := core.NewQuery().Add(
	// 	core.Aggs(
	// 		core.AggItem{
	// 			Name: "by_network",
	// 			Body: core.Item{
	// 				"terms": core.Item{
	// 					"field": "network.keyword",
	// 					"size":  core.MaxQuerySize,
	// 				},
	// 				"aggs": core.Item{
	// 					"last": core.TopHits(1, "level", "desc"),
	// 				},
	// 			},
	// 		},
	// 	),
	// ).Zero()

	// var response getLastBlocksResponse
	// if err := storage.es.Query([]string{models.DocBlocks}, query, &response); err != nil {
	// 	return nil, err
	// }

	// buckets := response.Agg.ByNetwork.Buckets
	// blocks := make([]block.Block, len(buckets))
	// for i := range buckets {
	// 	var block block.Block
	// 	if err := json.Unmarshal(buckets[i].Last.Hits.Hits[0].Source, &block); err != nil {
	// 		return nil, err
	// 	}
	// 	blocks[i] = block
	// }
	return nil, nil
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	builder := core.NewBuilder().SelectAll(models.DocBlocks).And(
		core.NewEq("chain_id", chainID),
	).One()

	var block block.Block
	err := storage.db.GetOne(builder, &block)
	return block.Network, err
}
