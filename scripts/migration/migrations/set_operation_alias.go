package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar"
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
func (m *SetOperationAlias) Do(ctx *Context) error {
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

	for i := range operations {
		bar.Add(1)
		h.SetOperationAliases(aliases, &operations[i])

		if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operations[i].ID, operations[i]); err != nil {
			fmt.Print("\033[2K\r")
			return err
		}
	}

	fmt.Print("\033[2K\r")
	logger.Info("[%s] done. Total operations: %d", m.Network, len(operations))

	return nil
}
