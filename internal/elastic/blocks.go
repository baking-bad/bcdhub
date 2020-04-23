package elastic

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
)

// CurrentState - returns current indexer state for network
func (e *Elastic) CurrentState(network string) (block models.Block, err error) {
	block.Network = network

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Sort("level", "desc").One()

	r, err := e.query([]string{DocBlocks}, query)
	if err != nil {
		if strings.Contains(err.Error(), IndexNotFoundError) {
			return block, nil
		}
		return
	}

	if r.Get("hits.total.value").Int() == 0 {
		return block, nil
	}
	hit := r.Get("hits.hits.0")
	block.ParseElasticJSON(hit)
	return
}

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

	r, err := e.query([]string{DocBlocks}, query)
	if err != nil {
		return
	}

	if r.Get("hits.total.value").Int() == 0 {
		return block, fmt.Errorf("%s: block in %s at level %d", RecordNotFound, network, level)
	}
	hit := r.Get("hits.hits.0")
	block.ParseElasticJSON(hit)
	return
}

// GetAllStates - return last block for all networks
func (e *Elastic) GetAllStates() ([]models.Block, error) {
	query := newQuery().Add(
		aggs("by_network", qItem{
			"terms": qItem{
				"field": "network.keyword",
				"size":  maxQuerySize,
			},
			"aggs": qItem{
				"last": topHits(1, "level", "desc"),
			},
		}),
	).Zero()

	response, err := e.query([]string{DocBlocks}, query)
	if err != nil {
		return nil, err
	}

	hits := response.Get("aggregations.by_network.buckets.#.last.hits.hits.0").Array()
	blocks := make([]models.Block, len(hits))
	for i, hit := range hits {
		blocks[i].ParseElasticJSON(hit)
	}
	return blocks, nil
}
