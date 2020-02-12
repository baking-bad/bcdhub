package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Tags -
type Tags struct {
	*DefaultMetric
}

// NewTags -
func NewTags(weight float64) *Tags {
	return &Tags{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Tags) Compute(a, b models.Contract) float64 {
	sum := 0.0
	if len(a.Tags) == 0 && len(b.Tags) == 0 {
		return m.Weight
	}

	for i := range a.Tags {
		found := false

		for j := range b.Tags {
			if b.Tags[j] == a.Tags[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	return sum * m.Weight / float64(len(a.Tags)+len(b.Tags))
}
