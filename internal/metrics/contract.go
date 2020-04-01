package metrics

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/classification/functions"
	clmetrics "github.com/baking-bad/bcdhub/internal/classification/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/google/uuid"
)

// SetContractAlias -
func (h *Handler) SetContractAlias(aliases map[string]string, c *models.Contract) {
	c.Alias = aliases[c.Address]
}

// SetContractStats -
func (h *Handler) SetContractStats(c *models.Contract) error {
	stats, err := h.ES.GetContractStats(c.Address, c.Network)
	if err != nil {
		return err
	}
	c.TxCount = stats.TxCount
	c.LastAction = models.BCDTime{
		Time: stats.LastAction,
	}
	c.SumTxAmount = stats.SumTxAmount
	c.MedianConsumedGas = stats.MedianConsumedGas

	return nil
}

// SetContractProjectID -
func (h *Handler) SetContractProjectID(c *models.Contract) error {
	buckets, err := h.ES.GetLastProjectContracts()
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

	return strings.ReplaceAll(uuid.New().String(), "-", ""), nil
}

var model = []clmetrics.Metric{
	clmetrics.NewBool("Manager"),
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
			return false, fmt.Errorf("Invalid metric weights. Check sum of weight is not equal 1")
		}
	}

	clf := functions.NewLinearSVC()
	res := clf.Predict(features)
	// log.Printf("%s -> %s [%d]", a.Address, b.Address, res)
	return res == 1, nil
}
