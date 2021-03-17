package reindexer

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetDateHistogram -
func (r *Reindexer) GetDateHistogram(period string, opts ...models.HistogramOption) ([][]int64, error) {
	return make([][]int64, 0), nil
}
