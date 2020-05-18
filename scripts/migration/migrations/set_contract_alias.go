package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// SetContractAlias - migration that set alias from db to contracts in choosen network
type SetContractAlias struct {
	Network string
}

// Key -
func (m *SetContractAlias) Key() string {
	return "contract_alias"
}

// Description -
func (m *SetContractAlias) Description() string {
	return "set alias from db to contracts in choosen network"
}

// Do - migrate function
func (m *SetContractAlias) Do(ctx *config.Context) error {
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
