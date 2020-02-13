package main

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/cmd/finder/metrics"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
)

var model = []metrics.Metric{
	metrics.NewManager(0.15),
	metrics.NewTags(0.1),
	metrics.NewFailStrings(0.1),
	metrics.NewAnnotations(0.1),
	metrics.NewLanguage(0.05),
	// metrics.NewPrimitives(0.05),
	metrics.NewEntrypoints(0.05),
	metrics.NewFingerprintLength(0.05, "parameter"),
	metrics.NewFingerprintLength(0.05, "storage"),
	metrics.NewFingerprintLength(0.05, "code"),
	metrics.NewFingerprint(0.1, "parameter"),
	metrics.NewFingerprint(0.1, "storage"),
	metrics.NewFingerprint(0.1, "code"),
}

func compare(a, b models.Contract) (bool, error) {
	sum := 0.0
	for i := range model {
		sum += model[i].Compute(a, b)
		if sum > 1 {
			return false, fmt.Errorf("Invalid metric weights. Check sum of weight is not equal 1")
		}
	}

	// log.Printf("%s -> %s [%.3f]", a.Address, b.Address, sum)
	return 0.85 <= sum, nil
}
