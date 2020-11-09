package elastic

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
)

// GetBlock -
func (e *Elastic) GetBlock(network string, level int64) (block models.Block, err error) {
	block.Network = network

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				term("level", level),
			),
		),
	).One()

	var response SearchResponse
	if err = e.query([]string{DocBlocks}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return block, NewRecordNotFoundError(DocBlocks, "", query)
	}

	err = json.Unmarshal(response.Hits.Hits[0].Source, &block)
	return
}

// GetLastBlock - returns current indexer state for network
func (e *Elastic) GetLastBlock(network string) (block models.Block, err error) {
	block.Network = network

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Sort("level", "desc").One()

	var response SearchResponse
	if err = e.query([]string{DocBlocks}, query, &response); err != nil {
		if strings.Contains(err.Error(), IndexNotFoundError) {
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

type getLastBlocksResponse struct {
	Agg struct {
		ByNetwork struct {
			Buckets []struct {
				Last struct {
					Hits HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"by_network"`
	} `json:"aggregations"`
}

// GetLastBlocks - return last block for all networks
func (e *Elastic) GetLastBlocks() ([]models.Block, error) {
	query := newQuery().Add(
		aggs(
			aggItem{
				"by_network", qItem{
					"terms": qItem{
						"field": "network.keyword",
						"size":  maxQuerySize,
					},
					"aggs": qItem{
						"last": topHits(1, "level", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getLastBlocksResponse
	if err := e.query([]string{DocBlocks}, query, &response); err != nil {
		return nil, err
	}

	buckets := response.Agg.ByNetwork.Buckets
	blocks := make([]models.Block, len(buckets))
	for i := range buckets {
		var block models.Block
		if err := json.Unmarshal(buckets[i].Last.Hits.Hits[0].Source, &block); err != nil {
			return nil, err
		}
		blocks[i] = block
	}
	return blocks, nil
}

// GetNetworkAlias -
func (e *Elastic) GetNetworkAlias(chainID string) (string, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("chain_id", chainID),
			),
		),
	).One()

	var response SearchResponse
	if err := e.query([]string{DocBlocks}, query, &response); err != nil {
		return "", err
	}

	if response.Hits.Total.Value == 0 {
		return "", NewRecordNotFoundError(DocBlocks, "", query)
	}

	var block models.Block
	err := json.Unmarshal(response.Hits.Hits[0].Source, &block)
	return block.Network, err
}
