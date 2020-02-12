package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Manager -
type Manager struct {
	*DefaultMetric
}

// NewManager -
func NewManager(weight float64) *Manager {
	return &Manager{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Manager) Compute(a, b models.Contract) float64 {
	if a.Manager == b.Manager {
		return m.Weight
	}
	return 0
}
