package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// SearchByText -
func (r *Reindexer) SearchByText(text string, offset int64, fields []string, filters map[string]interface{}, group bool) (models.Result, error) {
	return models.Result{}, nil
}
