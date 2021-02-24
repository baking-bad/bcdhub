package protocol

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// GetProtocol - returns current protocol for `network` and `level` (`hash` is optional, leave empty string for default)
func (storage *Storage) GetProtocol(network, hash string, level int64) (p protocol.Protocol, err error) {
	filters := []core.Item{
		core.Match("network", network),
	}
	if level > -1 {
		filters = append(filters,
			core.Range("start_level", core.Item{
				"lte": level,
			}),
		)
	}
	if hash != "" {
		filters = append(filters,
			core.Match("hash", hash),
		)
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(filters...),
		),
	).Sort("start_level", "desc").One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocProtocol}, query, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		err = core.NewRecordNotFoundError(models.DocProtocol, "")
		return
	}
	p.ID = response.Hits.Hits[0].ID
	err = json.Unmarshal(response.Hits.Hits[0].Source, &p)
	return
}

// GetSymLinks - returns list of symlinks in `network` after `level`
func (storage *Storage) GetSymLinks(network string, level int64) (map[string]struct{}, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Range("start_level", core.Item{
					"gt": level,
				}),
			),
		),
	).Sort("start_level", "desc").All()
	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocProtocol}, query, &response); err != nil {
		return nil, err
	}

	symMap := make(map[string]struct{})
	for _, hit := range response.Hits.Hits {
		var p protocol.Protocol
		if err := json.Unmarshal(hit.Source, &p); err != nil {
			return nil, err
		}
		symMap[p.SymLink] = struct{}{}
	}
	return symMap, nil
}
