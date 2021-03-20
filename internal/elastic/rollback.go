package elastic

import "github.com/baking-bad/bcdhub/internal/search"

// Rollback -
func (e *Elastic) Rollback(network string, level int64) error {
	return e.DeleteByLevelAndNetwork(search.Indices, network, level)
}
