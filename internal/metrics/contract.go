package metrics

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
)

// SetScriptProjectID -
func SetScriptProjectID(scripts contract.ScriptRepository, c *contract.Script, chunk []contract.Script) error {
	projectID := getContractProjectID(*c, chunk)
	if projectID != "" {
		c.ProjectID = types.NewNullString(&projectID)
		return nil
	}

	var offset int
	size := 100
	var end bool
	for !end {
		buckets, err := scripts.GetScripts(size, offset)
		if err != nil {
			return err
		}
		end = len(buckets) < size

		if !end {
			projectID := getContractProjectID(*c, buckets)
			if projectID != "" {
				c.ProjectID = types.NewNullString(&projectID)
				return nil
			}
		}

		offset += size
	}

	projectID = helpers.GenerateID()
	c.ProjectID = types.NewNullString(&projectID)
	return nil
}

func getContractProjectID(c contract.Script, buckets []contract.Script) string {
	for i := len(buckets) - 1; i > -1; i-- {
		if buckets[i].ProjectID.Valid && compare(c, buckets[i]) {
			return buckets[i].ProjectID.String()
		}
	}

	return ""
}

var precomputedMetrics = []clmetrics.Metric{
	clmetrics.NewBinMask("Tags"),
	clmetrics.NewArray("FailStrings"),
	clmetrics.NewArray("Annotations"),
	clmetrics.NewArray("Entrypoints"),
	clmetrics.NewFingerprintLength("parameter"),
	clmetrics.NewFingerprintLength("storage"),
	clmetrics.NewFingerprintLength("code"),
}

var fingerprintMetrics = []clmetrics.Metric{
	clmetrics.NewFingerprint("parameter"),
	clmetrics.NewFingerprint("storage"),
	clmetrics.NewFingerprint("code"),
}

func compare(a, b contract.Script) bool {
	features := make([]float64, len(precomputedMetrics))

	for i := range precomputedMetrics {
		f := precomputedMetrics[i].Compute(a, b)
		features[i] = f.Value
	}

	clf := functions.NewPrecomputedLinearSVC()
	res := clf.Predict(features)
	if res != 1 {
		return false
	}

	for i := range fingerprintMetrics {
		f := fingerprintMetrics[i].Compute(a, b)
		features = append(features, f.Value)
	}

	fullClf := functions.NewLinearSVC()
	res = fullClf.Predict(features)
	return res == 1
}
