package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar"
)

// SetContractAliasMigration - migration that set alias from db to contracts in choosen network
type SetContractAliasMigration struct {
	Network string
}

// Do - migrate function
func (m *SetContractAliasMigration) Do(ctx *Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	filter := make(map[string]interface{})
	filter["network"] = m.Network

	contracts, err := ctx.ES.GetContracts(filter)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts in %s", len(contracts), m.Network)

	aliases, err := ctx.DB.GetAliasesMap(m.Network)
	if err != nil {
		return err
	}

	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false))

	for i := range contracts {
		bar.Add(1)
		h.SetContractAlias(aliases, &contracts[i])

		if _, err := ctx.ES.UpdateDoc(elastic.DocContracts, contracts[i].ID, contracts[i]); err != nil {
			fmt.Print("\033[2K\r")
			return err
		}
	}

	fmt.Print("\033[2K\r")
	logger.Info("[%s] done. Total contracts: %d", m.Network, len(contracts))

	return nil
}
