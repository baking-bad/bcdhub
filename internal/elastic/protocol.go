package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
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
	response, err := e.query([]string{DocProtocol}, query)
	if err != nil {
		return
	}
	if response.Get("hits.total.value").Int() == 0 {
		err = fmt.Errorf("Couldn't find a protocol for %s (hash = %s) at level %d", network, hash, level)
		return
	}
	hit := response.Get("hits.hits.0")
	p.ParseElasticJSON(hit)
	return
}

// GetProtocols - returns all protocols
func (e *Elastic) GetProtocols() ([]models.Protocol, error) {
	query := newQuery().Sort("start_level", "desc").All()
	response, err := e.query([]string{DocProtocol}, query)
	if err != nil {
		return nil, err
	}

	protocols := make([]models.Protocol, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var p models.Protocol
		p.ParseElasticJSON(hit)
		protocols = append(protocols, p)
	}
	return protocols, nil
}

// GetProtocolsByNetwork - returns all protocols by `network`
func (e *Elastic) GetProtocolsByNetwork(network string) ([]models.Protocol, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Sort("start_level", "desc").All()
	response, err := e.query([]string{DocProtocol}, query)
	if err != nil {
		return nil, err
	}

	protocols := make([]models.Protocol, 0)
	for _, hit := range response.Get("hits.hits").Array() {
		var p models.Protocol
		p.ParseElasticJSON(hit)
		protocols = append(protocols, p)
	}
	return protocols, nil
}

// GetSymLinks - returns list of symlinks in `network` after `level`
func (e *Elastic) GetSymLinks(network string, level int64) (map[string]bool, error) {
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
	response, err := e.query([]string{DocProtocol}, query)
	if err != nil {
		return nil, err
	}
	symMap := make(map[string]bool)
	for _, hit := range response.Get("hits.hits").Array() {
		symLink := hit.Get("_source.sym_link").String()
		symMap[symLink] = true
	}
	return symMap, nil
}
