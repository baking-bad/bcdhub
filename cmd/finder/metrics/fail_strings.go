package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// FailStrings -
type FailStrings struct {
	*DefaultMetric
}

// NewFailStrings -
func NewFailStrings(weight float64) *FailStrings {
	return &FailStrings{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *FailStrings) Compute(a, b models.Contract) float64 {
	sum := 0.0
	if len(a.FailStrings) == 0 && len(b.FailStrings) == 0 {
		return m.Weight
	}

	for i := range a.FailStrings {
		found := false

		for j := range b.FailStrings {
			if b.FailStrings[j] == a.FailStrings[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	return sum * m.Weight / float64(len(a.FailStrings)+len(b.FailStrings))
}
