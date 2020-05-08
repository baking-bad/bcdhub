package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// SetOperationAlias - migration that set source or destination alias from db to operations in choosen network
type SetOperationAlias struct {
	Network string
}

// Description -
func (m *SetOperationAlias) Description() string {
	return "set source or destination alias from db to operations in choosen network"
}

// Do - migrate function
func (m *SetOperationAlias) Do(ctx *config.Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	operations, err := ctx.ES.GetAllOperations(m.Network)
	if err != nil {
		return err
	}

	logger.Info("Found %d operations in %s", len(operations), m.Network)

	aliases, err := ctx.DB.GetAliasesMap(m.Network)
	if err != nil {
		return err
	}

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false))

	var changed int64
	var bulk []elastic.Identifiable

	for i := range operations {
		bar.Add(1)

		found := h.SetOperationAliases(aliases, &operations[i])

		if found {
			changed++
			bulk = append(bulk, operations[i])
		}

		if len(bulk) == 1000 || (i == len(operations)-1 && len(bulk) > 0) {
			if err := ctx.ES.BulkUpdate("operation", bulk); err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
			bulk = bulk[:0]
		}
	}

	fmt.Print("\033[2K\r")
	logger.Info("[%s] done. Total operations: %d. Changed: %d", m.Network, len(operations), changed)

	return nil
}
