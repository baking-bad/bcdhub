package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetBMDProtocol - migration that set `Protocol` at big map diff
type SetBMDProtocol struct{}

// Description -
func (m *SetBMDProtocol) Description() string {
	return "set `Protocol` at big map diff"
}

// Do - migrate function
func (m *SetBMDProtocol) Do(ctx *config.Context) error {
	allBMD, err := ctx.ES.GetAllBigMapDiff()
	if err != nil {
		return err
	}
	logger.Info("Found %d unique operations with big map diff", len(allBMD))

	bar := progressbar.NewOptions(len(allBMD), progressbar.OptionSetPredictTime(false))
	ops := make(map[string]string)
	var lastIdx int

	for i := range allBMD {
		bar.Add(1)

		proto, ok := ops[allBMD[i].OperationID]
		if !ok {
			operation, err := ctx.ES.GetByID(elastic.DocOperations, allBMD[i].OperationID)
			if err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
			proto = operation.Get("_source.protocol").String()
		}
		allBMD[i].Protocol = proto

		if (i%1000 == 0 || i == len(allBMD)-1) && i > 0 {
			updates := make([]elastic.Identifiable, len(allBMD[lastIdx:i]))
			for j := range allBMD[lastIdx:i] {
				updates[j] = allBMD[lastIdx:i][j]
			}
			if err := ctx.ES.BulkUpdate("bigmapdiff", updates); err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
			lastIdx = i
		}
	}

	fmt.Print("\033[2K\r")
	logger.Info("Done.")

	return nil
}
