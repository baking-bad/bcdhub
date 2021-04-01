package elastic

import "github.com/baking-bad/bcdhub/internal/search"

// Rollback -
func (e *Elastic) Rollback(network string, level int64) error {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network),
				Range("level", Item{"gt": level}),
			),
		),
	)
	end := false

	for !end {
		response, err := e.deleteWithQuery(search.Indices, query)
		if err != nil {
			return err
		}

		end = response.VersionConflicts == 0
	}
	return nil
}
