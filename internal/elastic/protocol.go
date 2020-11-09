package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// GetProtocol - returns current protocol for `network` and `level` (`hash` is optional, leave empty string for default)
func (e *Elastic) GetProtocol(network, hash string, level int64) (p models.Protocol, err error) {
	filters := []qItem{
		matchQ("network", network),
	}
	if level > -1 {
		filters = append(filters,
			rangeQ("start_level", qItem{
				"lte": level,
			}),
		)
	}
	if hash != "" {
		filters = append(filters,
			matchQ("hash", hash),
		)
	}

	query := newQuery().Query(
		boolQ(
			filter(filters...),
		),
	).Sort("start_level", "desc").One()

	var response SearchResponse
	if err = e.query([]string{DocProtocol}, query, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		err = errors.Errorf("Couldn't find a protocol for %s (hash = %s) at level %d", network, hash, level)
		return
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &p)
	return
}

// GetSymLinks - returns list of symlinks in `network` after `level`
func (e *Elastic) GetSymLinks(network string, level int64) (map[string]struct{}, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("start_level", qItem{
					"gt": level,
				}),
			),
		),
	).Sort("start_level", "desc").All()
	var response SearchResponse
	if err := e.query([]string{DocProtocol}, query, &response); err != nil {
		return nil, err
	}

	symMap := make(map[string]struct{})
	for _, hit := range response.Hits.Hits {
		var p models.Protocol
		if err := json.Unmarshal(hit.Source, &p); err != nil {
			return nil, err
		}
		symMap[p.SymLink] = struct{}{}
	}
	return symMap, nil
}
