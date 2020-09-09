package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
)

// SetOperationAllocatedBurned -
type SetOperationAllocatedBurned struct{}

// Key -
func (m *SetOperationAllocatedBurned) Key() string {
	return "set_allocated_burned"
}

// Description -
func (m *SetOperationAllocatedBurned) Description() string {
	return "set field `AllocatedDestinationContractBurned` in `Operation` model"
}

// Do - migrate function
func (m *SetOperationAllocatedBurned) Do(ctx *config.Context) error {
	logger.Info("Fetching operations...")
	operations, err := ctx.ES.GetOperations(map[string]interface{}{
		"kind":   "origination",
		"status": "applied",
	}, 0, false)
	if err != nil {
		return err
	}

	logger.Info("Fetching protocols...")
	protocols := make([]models.Protocol, 0)
	if err := ctx.ES.GetAll(&protocols); err != nil {
		return err
	}

	protoMap := make(map[string]models.Protocol)
	for _, p := range protocols {
		protoMap[p.Hash] = p
	}

	logger.Info("Computing allocation burn...")
	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
	updated := make([]elastic.Model, len(operations))
	for i := range operations {
		bar.Add(1) //nolint
		if operations[i].Result.AllocatedDestinationContract {
			p, ok := protoMap[operations[i].Protocol]
			if !ok {
				return errors.Errorf("Unknown protocol: %s", operations[i].Protocol)
			}
			operations[i].SetAllocationBurn(p.Constants)
		}
		updated[i] = &operations[i]
	}

	return ctx.ES.BulkUpdate(updated)
}
