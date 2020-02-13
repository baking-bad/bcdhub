package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Annotations -
type Annotations struct {
	*DefaultMetric
}

// NewAnnotations -
func NewAnnotations(weight float64) *Annotations {
	return &Annotations{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Annotations) Compute(a, b models.Contract) float64 {
	sum := 0.0
	if len(a.Annotations) == 0 && len(b.Annotations) == 0 {
		return m.Weight
	}
	if len(a.Annotations) == 0 || len(b.Annotations) == 0 {
		return 0
	}

	for i := range a.Annotations {
		found := false

		for j := range b.Annotations {
			if b.Annotations[j] == a.Annotations[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	return round(sum*m.Weight/float64(len(a.Annotations)+len(b.Annotations)), 6)
}
