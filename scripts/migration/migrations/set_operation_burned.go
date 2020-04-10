package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// SetOperationBurned - migration that set burned to operations in all networks
type SetOperationBurned struct{}

// Description -
func (m *SetOperationBurned) Description() string {
	return "set burned to operations in all networks"
}

// Do - migrate function
func (m *SetOperationBurned) Do(ctx *Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	for _, network := range []string{consts.Mainnet, consts.Zeronet, consts.Carthage, consts.Babylon} {
		operations, err := ctx.ES.GetAllOperations(network)
		if err != nil {
			return err
		}

		logger.Info("Found %d operations in %s", len(operations), network)

		bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false))

		for i := range operations {
			bar.Add(1)
			h.SetOperationBurned(&operations[i])
			if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operations[i].ID, operations[i]); err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
		}

		fmt.Print("\033[2K\r")
		logger.Info("[%s] done. Total operations: %d", network, len(operations))
	}
	return nil
}
