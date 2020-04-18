package elastic

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
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
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).One()

	r, err := e.query([]string{DocStates}, query)
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
	s.ParseElasticJSON(hit)
	return
}
