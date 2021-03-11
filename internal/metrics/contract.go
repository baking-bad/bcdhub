package metrics

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
)

// SetContractAlias -
func (h *Handler) SetContractAlias(c *contract.Contract, aliases map[string]string) (bool, error) {
	var changed bool

	if c.Network != consts.Mainnet || len(aliases) == 0 {
		return false, nil
	}

	if c.Alias != "" && (c.Delegate != "" || c.DelegateAlias != "") {
		return false, nil
	}

	if alias, ok := aliases[c.Address]; ok && c.Alias == "" {
		c.Alias = alias
		changed = true
	}

	if alias, ok := aliases[c.Delegate]; c.Delegate != "" && c.DelegateAlias == "" && ok {
		c.DelegateAlias = alias
		changed = true
	}

	return changed, nil
}

// UpdateContractStats -
func (h *Handler) UpdateContractStats(c *contract.Contract) error {
	count, err := h.Migrations.Count(c.Network, c.Address)
	if err != nil {
		return err
	}
	c.MigrationsCount = count
	return nil
}

// SetContractProjectID -
func (h *Handler) SetContractProjectID(c *contract.Contract) error {
	buckets, err := h.Contracts.GetProjectsLastContract(c)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			c.ProjectID = helpers.GenerateID()
			return nil
		}
		return err
	}

	c.ProjectID = getContractProjectID(*c, buckets)

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
