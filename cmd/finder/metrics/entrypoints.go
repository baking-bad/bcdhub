package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Entrypoints -
type Entrypoints struct {
	*DefaultMetric
}

// NewEntrypoints -
func NewEntrypoints(weight float64) *Entrypoints {
	return &Entrypoints{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Entrypoints) Compute(a, b models.Contract) float64 {
	sum := 0.0
	if len(a.Entrypoints) == 0 && len(b.Entrypoints) == 0 {
		return m.Weight
	}
	if len(a.Entrypoints) == 0 || len(b.Entrypoints) == 0 {
		return 0
	}

	for i := range a.Entrypoints {
		found := false

		for j := range b.Entrypoints {
			if b.Entrypoints[j] == a.Entrypoints[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	return round(sum*m.Weight/float64(len(a.Entrypoints)+len(b.Entrypoints)), 6)
}
