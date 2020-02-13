package metrics

import (
	"encoding/hex"
	"math"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

const (
	levensteinThreashold = 100
)

// Fingerprint -
type Fingerprint struct {
	*DefaultMetric

	Section string
}

// NewFingerprint -
func NewFingerprint(weight float64, section string) *Fingerprint {
	return &Fingerprint{
		DefaultMetric: &DefaultMetric{
			Weight: weight,
		},
		Section: section,
	}
}

// Compute -
func (m *Fingerprint) Compute(a, b models.Contract) float64 {
	if a.Fingerprint == nil || b.Fingerprint == nil {
		return 0.0
	}

	var x, y []byte
	if m.Section == "parameter" {
		x, _ = hex.DecodeString(a.Fingerprint.Parameter)
		y, _ = hex.DecodeString(b.Fingerprint.Parameter)
	} else if m.Section == "storage" {
		x, _ = hex.DecodeString(a.Fingerprint.Storage)
		y, _ = hex.DecodeString(b.Fingerprint.Storage)
	} else if m.Section == "code" {
		x, _ = hex.DecodeString(a.Fingerprint.Code)
		y, _ = hex.DecodeString(b.Fingerprint.Code)
	} else {
		return 0.0
	}

	dist := float64(distance(x, y))
	maxLen := math.Max(float64(len(x)), float64(len(y)))
	val := 1. - math.Pow(dist/maxLen, 1.25)
	return round(val*m.Weight, 6)
}

func distance(a, b []byte) int {
	if len(a) == 0 {
		return len(b)
	}

	if len(b) == 0 {
		return len(a)
	}

	if len(a) == len(b) {
		eq := true
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				eq = false
				break
			}
		}
		if eq {
			return 0
		}
	}

	// swap to save some memory O(min(a,b)) instead of O(a)
	if len(a) > len(b) {
		a, b = b, a
	}

	lenA := len(a)
	lenB := len(b)

	x := make([]uint16, lenA+1)
	for i := 1; i < len(x); i++ {
		x[i] = uint16(i)
	}

	// make a dummy bounds check to prevent the 2 bounds check down below.
	// The one inside the loop is particularly costly.
	_ = x[lenA]

	// fill in the rest
	for i := 1; i <= lenB; i++ {
		prev := uint16(i)
		var current uint16

		// log.Printf("b[%d] = %s", i-1, string(b[i-1]))
		for j := 1; j <= lenA; j++ {
			// log.Printf("a[%d] = %s", j-1, string(a[j-1]))
			if b[i-1] == a[j-1] {
				current = x[j-1] // match
			} else {
				current = min(min(x[j-1]+1, prev+1), x[j]+1)
			}
			x[j-1] = prev
			prev = current
		}
		x[lenA] = prev
	}
	return int(x[lenA])
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

// FingerprintLength -
type FingerprintLength struct {
	*DefaultMetric

	Section string
}

// NewFingerprintLength -
func NewFingerprintLength(weight float64, section string) *FingerprintLength {
	return &FingerprintLength{
		DefaultMetric: &DefaultMetric{
			Weight: weight,
		},
		Section: section,
	}
}

// Compute -
func (m *FingerprintLength) Compute(a, b models.Contract) float64 {
	if a.Fingerprint == nil || b.Fingerprint == nil {
		return 0.0
	}
	var x, y string
	if m.Section == "parameter" {
		x = a.Fingerprint.Parameter
		y = b.Fingerprint.Parameter
	} else if m.Section == "storage" {
		x = a.Fingerprint.Storage
		y = b.Fingerprint.Storage
	} else if m.Section == "code" {
		x = a.Fingerprint.Code
		y = b.Fingerprint.Code
	} else {
		return 0.0
	}

	lx := float64(len(x))
	ly := float64(len(y))
	sum := float64(math.Min(lx, ly) / math.Max(lx, ly))
	return round(sum*m.Weight, 6)
}
