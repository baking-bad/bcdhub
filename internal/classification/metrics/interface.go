package metrics

import "github.com/baking-bad/bcdhub/internal/models/contract"

// Metric -
type Metric interface {
	Compute(a, b contract.Contract) Feature
}

// Feature -
type Feature struct {
	Value float64
	Name  string
}
