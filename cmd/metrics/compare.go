package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	"github.com/baking-bad/bcdhub/internal/classification/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
)

var model = []metrics.Metric{
	metrics.NewBool("Manager"),
	metrics.NewArray("Tags"),
	metrics.NewArray("FailStrings"),
	metrics.NewArray("Annotations"),
	metrics.NewBool("Language"),
	metrics.NewArray("Entrypoints"),
	metrics.NewFingerprintLength("parameter"),
	metrics.NewFingerprintLength("storage"),
	metrics.NewFingerprintLength("code"),
	metrics.NewFingerprint("parameter"),
	metrics.NewFingerprint("storage"),
	metrics.NewFingerprint("code"),
}

func compare(a, b models.Contract) (bool, error) {
	sum := 0.0
	features := make([]float64, len(model))

	for i := range model {
		f := model[i].Compute(a, b)
		features[i] = f.Value
		if sum > 1 {
			return false, fmt.Errorf("Invalid metric weights. Check sum of weight is not equal 1")
		}
	}

	clf := functions.NewLinearSVC()
	res := clf.Predict(features)
	// log.Printf("%s -> %s [%d]", a.Address, b.Address, res)
	return res == 1, nil
}
