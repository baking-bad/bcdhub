package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetTotalWithdrawn - migration that set total_withdrawn to contracts in all networks
type SetTotalWithdrawn struct {
	Network string
}

// Key -
func (m *SetTotalWithdrawn) Key() string {
	return "total_withdrawn"
}

// Description -
func (m *SetTotalWithdrawn) Description() string {
	return "set total_withdrawn to contracts in all networks"
}

// Do - migrate function
func (m *SetTotalWithdrawn) Do(ctx *config.Context) error {
	for _, network := range ctx.Config.Migrations.Networks {
		filter := make(map[string]interface{})
		filter["network"] = network

		contracts, err := ctx.ES.GetContracts(filter)
		if err != nil {
			return err
		}

		logger.Info("Found %d contracts in %s", len(contracts), network)

		bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false))

		for i, c := range contracts {
			bar.Add(1)

			totalWithdrawn, err := ctx.ES.GetContractWithdrawn(c.Address, c.Network)
			if err != nil {
				fmt.Print("\033[2K\r")
				return err
			}

			if totalWithdrawn > 0 {
				contracts[i].TotalWithdrawn = totalWithdrawn

				if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, contracts[i].ID, contracts[i]); err != nil {
					fmt.Print("\033[2K\r")
					return err
				}
			}
		}

		fmt.Print("\033[2K\r")
		logger.Info("[%s] done. Total contracts: %d", network, len(contracts))
	}
	return nil
}
