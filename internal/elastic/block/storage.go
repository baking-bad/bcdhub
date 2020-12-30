package block

import (
	"encoding/json"
	"strings"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// Get -
func (storage *Storage) Get(network string, level int64) (block block.Block, err error) {
	block.Network = network

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Term("level", level),
			),
		),
	).One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocBlocks}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return block, core.NewRecordNotFoundError(models.DocBlocks, "")
	}

	err = json.Unmarshal(response.Hits.Hits[0].Source, &block)
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last(network string) (block block.Block, err error) {
	block.Network = network

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
		),
	).Sort("level", "desc").One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocBlocks}, query, &response); err != nil {
		if strings.Contains(err.Error(), consts.IndexNotFoundError) {
			return block, nil
		}
		return
	}

	if response.Hits.Total.Value == 0 {
		return block, nil
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &block)
	return
}

// LastByNetworks - return last block for all networks
func (storage *Storage) LastByNetworks() ([]block.Block, error) {
	query := core.NewQuery().Add(
		core.Aggs(
			core.AggItem{
				Name: "by_network",
				Body: core.Item{
					"terms": core.Item{
						"field": "network.keyword",
						"size":  core.MaxQuerySize,
					},
					"aggs": core.Item{
						"last": core.TopHits(1, "level", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getLastBlocksResponse
	if err := storage.es.Query([]string{models.DocBlocks}, query, &response); err != nil {
		return nil, err
	}

	buckets := response.Agg.ByNetwork.Buckets
	blocks := make([]block.Block, len(buckets))
	for i := range buckets {
		var block block.Block
		if err := json.Unmarshal(buckets[i].Last.Hits.Hits[0].Source, &block); err != nil {
			return nil, err
		}
		blocks[i] = block
	}
	return blocks, nil
}

// GetNetworkAlias -
func (storage *Storage) GetNetworkAlias(chainID string) (string, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("chain_id", chainID),
			),
		),
	).One()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocBlocks}, query, &response); err != nil {
		return "", err
	}

	if response.Hits.Total.Value == 0 {
		return "", core.NewRecordNotFoundError(models.DocBlocks, "")
	}

	var block block.Block
	err := json.Unmarshal(response.Hits.Hits[0].Source, &block)
	return block.Network, err
}
