package metrics

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetContractAlias -
func (h *Handler) SetContractAlias(aliases map[string]string, c *models.Contract) bool {
	var changed bool

	if alias, ok := aliases[c.Address]; ok {
		c.Alias = alias
		changed = true
	}

	if alias, ok := aliases[c.Delegate]; c.Delegate != "" && ok {
		c.DelegateAlias = alias
		changed = true
	}

	return changed
}

// UpdateContractStats -
func (h *Handler) UpdateContractStats(c *models.Contract) error {
	migrationsStats, err := h.ES.GetContractMigrationStats(c.Network, c.Address)
	if err != nil {
		return err
	}
	c.MigrationsCount = migrationsStats.MigrationsCount
	return nil
}

// SetContractProjectID -
func (h *Handler) SetContractProjectID(c *models.Contract) error {
	buckets, err := h.ES.GetProjectsLastContract()
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			c.ProjectID = helpers.GenerateID()
			return nil
		}
		return err
	}

	c.ProjectID = getContractProjectID(*c, buckets)

	return nil
}

func getContractProjectID(c models.Contract, buckets []models.Contract) string {
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

func compare(a, b models.Contract) bool {
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
