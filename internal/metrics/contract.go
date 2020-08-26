package metrics

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetContractAlias -
func (h *Handler) SetContractAlias(aliases map[string]string, c *models.Contract) {
	c.Alias = aliases[c.Address]

	if c.Delegate != "" {
		c.DelegateAlias = aliases[c.Delegate]
	}
}

// SetContractStats - TODO: update in a script
func (h *Handler) SetContractStats(op models.Operation, c *models.Contract) error {
	c.TxCount++
	c.LastAction = models.BCDTime{
		Time: op.Timestamp,
	}

	if op.Status != consts.Applied {
		return nil
	}

	if c.Address == op.Destination {
		c.Balance += op.Amount
	} else if c.Address == op.Source {
		c.TotalWithdrawn += op.Amount
		c.Balance -= op.Amount
	}

	return nil
}

// UpdateContractStats -
func (h *Handler) UpdateContractStats(c *models.Contract) error {
	stats, err := h.ES.RecalcContractStats(c.Network, c.Address)
	if err != nil {
		return err
	}
	migrationsStats, err := h.ES.GetContractMigrationStats(c.Network, c.Address)
	if err != nil {
		return err
	}

	c.TxCount = stats.TxCount
	c.LastAction = models.BCDTime{
		Time: stats.LastAction,
	}
	c.Balance = stats.Balance
	c.TotalWithdrawn = stats.TotalWithdrawn
	c.MigrationsCount = migrationsStats.MigrationsCount

	return nil
}

// SetContractProjectID -
func (h *Handler) SetContractProjectID(c *models.Contract) error {
	buckets, err := h.ES.GetProjectsLastContract()
	if err != nil {
		return err
	}
	projID, err := getContractProjectID(*c, buckets)
	if err != nil {
		return err
	}

	c.ProjectID = projID

	return nil
}

func getContractProjectID(c models.Contract, buckets []models.Contract) (string, error) {
	for i := len(buckets) - 1; i > -1; i-- {
		ok, err := compare(c, buckets[i])
		if err != nil {
			return "", err
		}

		if ok {
			return buckets[i].ProjectID, nil
		}
	}

	return helpers.GenerateID(), nil
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

func compare(a, b models.Contract) (bool, error) {
	sum := 0.0
	features := make([]float64, len(model))

	for i := range model {
		f := model[i].Compute(a, b)
		features[i] = f.Value
		if sum > 1 {
			return false, errors.Errorf("Invalid metric weights. Check sum of weight is not equal 1")
		}
	}

	clf := functions.NewLinearSVC()
	res := clf.Predict(features)
	// log.Printf("%s -> %s [%d]", a.Address, b.Address, res)
	return res == 1, nil
}
