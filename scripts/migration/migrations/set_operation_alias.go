package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// SetOperationAlias - migration that set source or destination alias from db to operations in choosen network
type SetOperationAlias struct {
	Network string
}

// Key -
func (m *SetOperationAlias) Key() string {
	return "operation_alias"
}

// Description -
func (m *SetOperationAlias) Description() string {
	return "set source or destination alias from db to operations in choosen network"
}

// Do - migrate function
func (m *SetOperationAlias) Do(ctx *config.Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	var operations []models.Operation
	if err := ctx.ES.GetByNetwork(m.Network, &operations); err != nil {
		return err
	}

	logger.Info("Found %d operations in %s", len(operations), m.Network)

	aliases, err := ctx.DB.GetAliasesMap(m.Network)
	if err != nil {
		return err
	}

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())

	var changed int64
	var bulk []elastic.Model

	for i := range operations {
		bar.Add(1)

		found := h.SetOperationAliases(aliases, &operations[i])

		if found {
			changed++
			bulk = append(bulk, &operations[i])
		}

		if len(bulk) == 1000 || (i == len(operations)-1 && len(bulk) > 0) {
			if err := ctx.ES.BulkUpdate(bulk); err != nil {
				return err
			}
			bulk = bulk[:0]
		}
	}

	logger.Info("[%s] done. Total operations: %d. Changed: %d", m.Network, len(operations), changed)

	return nil
}
