package elastic

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
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
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
				matchPhrase("type", typ),
			),
		),
	).One()

	r, err := e.query([]string{DocStates}, query)
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
	s.ParseElasticJSON(hit)
	return
}
