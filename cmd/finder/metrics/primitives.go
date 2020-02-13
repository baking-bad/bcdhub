package metrics

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Primitives -
type Primitives struct {
	*DefaultMetric
}

// NewPrimitives -
func NewPrimitives(weight float64) *Primitives {
	return &Primitives{
		&DefaultMetric{
			Weight: weight,
		},
	}
}

// Compute -
func (m *Primitives) Compute(a, b models.Contract) float64 {
	sum := 0.0
	if len(a.Primitives) == 0 && len(b.Primitives) == 0 {
		return m.Weight
	}
	if len(a.Primitives) == 0 || len(b.Primitives) == 0 {
		return 0
	}

	for i := range a.Primitives {
		found := false

		for j := range b.Primitives {
			if b.Primitives[j] == a.Primitives[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	return round(sum*m.Weight/float64(len(a.Primitives)+len(b.Primitives)), 6)
}
