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

func parseState(hit gjson.Result, s *models.State) {
	s.ID = hit.Get("_id").String()
	s.Network = network
	s.Type = hit.Get("_source.type").String()
	s.Level = hit.Get("_source.level").Int()
	s.Timestamp = hit.Get("_source.timestamp").Time()
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
	parseState(hit, &s)
	return
}
