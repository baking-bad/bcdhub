package metrics

import "github.com/baking-bad/bcdhub/internal/models"

// Metric -
type Metric interface {
	Compute(a, b models.Contract) Feature
}

// Feature -
type Feature struct {
	Value float64
	Name  string
}
