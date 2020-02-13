package elastic

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

func (e *Elastic) createState(network, typ string) (s models.State, err error) {
	s.Network = network
	s.Type = typ
	id, err := e.AddDocument(s, DocStates)
	if err != nil {
		return s, err
	}
	s.ID = id
	return
}

// CurrentState - returns current indexer state for network
func (e *Elastic) CurrentState(network, typ string) (s models.State, err error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"network": network,
						}},
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"type": typ,
						}},
				},
			},
		},
	}
	r, err := e.query(DocStates, query)
	if err != nil {
		if strings.Contains(err.Error(), IndexNotFoundError) {
			return e.createState(network, typ)
		}
		return
	}

	if r.Get("hits.total.value").Int() == 0 {
		return e.createState(network, typ)
	}
	hit := r.Get("hits.hits.0")
	s.ID = hit.Get("_id").String()
	s.Network = network
	s.Type = hit.Get("_source.type").String()
	s.Level = hit.Get("_source.level").Int()
	s.Timestamp = hit.Get("_source.timestamp").Time()
	return
}
