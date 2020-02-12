package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Metric -
type Metric interface {
	Compute(a, b models.Contract) float64
}

// DefaultMetric -
type DefaultMetric struct {
	Weight float64
}
