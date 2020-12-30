package metrics

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/contract"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
)

// SetContractAlias -
func (h *Handler) SetContractAlias(c *contract.Contract) (bool, error) {
	var changed bool

	if c.Alias != "" && ((c.Delegate != "" && c.DelegateAlias != "") || c.Delegate == "") {
		return false, nil
	}

	aliases, err := h.TZIP.GetAliasesMap(c.Network)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			err = nil
		}
		return changed, err
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
	buckets, err := h.Contracts.GetProjectsLastContract()
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
		if compare(c, buckets[i]) {
			return buckets[i].ProjectID
		}
	}

	return helpers.GenerateID()
}

var model = []clmetrics.Metric{
	clmetrics.NewManager(),
	clmetrics.NewArray("Tags"),
	clmetrics.NewArray("FailStrings"),
	clmetrics.NewArray("Annotations"),
	clmetrics.NewBool("Language"),
	clmetrics.NewArray("Entrypoints"),
	clmetrics.NewFingerprintLength("parameter"),
	clmetrics.NewFingerprintLength("storage"),
	clmetrics.NewFingerprintLength("code"),
	clmetrics.NewFingerprint("parameter"),
	clmetrics.NewFingerprint("storage"),
	clmetrics.NewFingerprint("code"),
}

func compare(a, b contract.Contract) bool {
	features := make([]float64, len(model))

	for i := range model {
		f := model[i].Compute(a, b)
		features[i] = f.Value
	}

	clf := functions.NewLinearSVC()
	res := clf.Predict(features)
	// log.Printf("%s -> %s [%d]", a.Address, b.Address, res)
	return res == 1
}
