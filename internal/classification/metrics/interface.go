package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Metric -
type Metric interface {
	Compute(a, b models.Contract) Feature
}

// Feature -
type Feature struct {
	Value float64
	Name  string
}
