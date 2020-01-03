package elastic

import (
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

func (e *Elastic) createState(network string) (s models.State, err error) {
	s.Network = network
	id, err := e.AddDocument(s, DocStates)
	if err != nil {
		return s, err
	}
	s.ID = id
	return
}

// CurrentState - returns current indexer state for network
func (e *Elastic) CurrentState(network string) (s models.State, err error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": map[string]interface{}{
					"term": map[string]interface{}{
						"network": network,
					},
				},
			},
		},
	}
	r, err := e.query(DocStates, query)
	if err != nil {
		if strings.Contains(err.Error(), IndexNotFoundError) {
			return e.createState(network)
		}
		return
	}
	if r.Get("hits.total.value").Int() == 0 {
		return e.createState(network)
	}
	hit := r.Get("hits.hits.0")
	s.ID = hit.Get("_id").String()
	s.Network = network
	s.Level = hit.Get("_source.level").Int()
	return
}
