package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetContractHash - migration that set `Hash` to contract
type SetContractHash struct{}

// Key -
func (m *SetContractHash) Key() string {
	return "set_contract_hash"
}

// Description -
func (m *SetContractHash) Description() string {
	return "set `Hash` to contract"
}

// Do - migrate function
func (m *SetContractHash) Do(ctx *config.Context) error {
	logger.Info("Start SetContractHash migration...")
	start := time.Now()

	for _, network := range ctx.Config.Migrations.Networks {
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return err
		}

		contracts, err := ctx.ES.GetContracts(map[string]interface{}{
			"network": network,
		})
		if err != nil {
			return err
		}
		state, err := ctx.ES.CurrentState(network)
		if err != nil {
			return err
		}

		logger.Info("Found %d contracts in %s", len(contracts), network)

		bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
		var lastIdx int
		for i := range contracts {
			bar.Add(1) //nolint
			if contracts[i].Hash != "" {
				continue
			}

			script, err := contractparser.GetContract(rpc, contracts[i].Address, contracts[i].Network, state.Protocol, ctx.Config.Share.Path, 0)
			if err != nil {
				return err
			}

			code := script.Get("code").Raw
			hash, err := contractparser.ComputeContractHash(code)
			if err != nil {
				return err
			}
			contracts[i].Hash = hash

			if (i%1000 == 0 || i == len(contracts)-1) && i > 0 {
				updates := make([]elastic.Model, len(contracts[lastIdx:i]))
				for j := range contracts[lastIdx:i] {
					updates[j] = &contracts[lastIdx:i][j]
				}
				if err := ctx.ES.BulkUpdate(updates); err != nil {
					return err
				}
				lastIdx = i
			}
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
