package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Language -
type Language struct {
	*DefaultMetric
}

// NewLanguage -
func NewLanguage(weight float64) *Language {
	return &Language{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Language) Compute(a, b models.Contract) float64 {
	if a.Language == b.Language {
		return m.Weight
	}
	return 0
}
