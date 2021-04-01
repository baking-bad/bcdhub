package metrics

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
)

// SetContractProjectID -
func (h *Handler) SetContractProjectID(c *contract.Contract, chunk []contract.Contract) error {
	var offset int64

	size := int64(25)

	var end bool
	for !end {
		buckets, err := h.Contracts.GetProjectsLastContract(*c, size, offset)
		if err != nil {
			return err
		}
		end = len(buckets) < int(size)

		if !end {
			c.ProjectID = getContractProjectID(*c, buckets)
			if c.ProjectID != "" {
				return nil
			}
		}

		offset += size
	}

	c.ProjectID = getContractProjectID(*c, chunk)
	if c.ProjectID != "" {
		return nil
	}
	c.ProjectID = helpers.GenerateID()
	return nil
}

func getContractProjectID(c contract.Contract, buckets []contract.Contract) string {
	for i := len(buckets) - 1; i > -1; i-- {
		if c.Hash == buckets[i].Hash {
			return buckets[i].ProjectID
		}
	}
	for i := len(buckets) - 1; i > -1; i-- {
		if compare(c, buckets[i]) {
			return buckets[i].ProjectID
		}
	}

	return helpers.GenerateID()
}

var precomputedMetrics = []clmetrics.Metric{
	clmetrics.NewManager(),
	clmetrics.NewArray("Tags"),
	clmetrics.NewArray("FailStrings"),
	clmetrics.NewArray("Annotations"),
	clmetrics.NewBool("Language"),
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

func compare(a, b contract.Contract) bool {
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
