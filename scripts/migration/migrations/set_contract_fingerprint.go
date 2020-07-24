package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// SetContractFingerprint - migration that set `Fingerprint` to contract
type SetContractFingerprint struct{}

// Key -
func (m *SetContractFingerprint) Key() string {
	return "set_contract_fingerprint"
}

// Description -
func (m *SetContractFingerprint) Description() string {
	return "set `Fingerprint` to contract"
}

// Do - migrate function
func (m *SetContractFingerprint) Do(ctx *config.Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	filter := make(map[string]interface{})

	contracts, err := ctx.ES.GetContracts(filter)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	state, err := ctx.ES.GetLastBlocks()
	if err != nil {
		return err
	}

	logger.Info("Getting current protocols...")
	protocols := map[string]string{}
	for i := range state {
		logger.Info("%s -> %s", state[i].Network, state[i].Protocol)
		protocols[state[i].Network] = state[i].Protocol
	}

	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	var lastIdx int
	for i := range contracts {
		bar.Add(1) //nolint

		rpc, err := ctx.GetRPC(contracts[i].Network)
		if err != nil {
			return err
		}

		protocol := protocols[contracts[i].Network]
		script, err := contractparser.GetContract(rpc, contracts[i].Address, contracts[i].Network, protocol, ctx.Config.Share.Path, 0)
		if err != nil {
			return err
		}

		if err := metrics.SetFingerprint(script, &contracts[i]); err != nil {
			return err
		}

		if err := h.SetContractProjectID(&contracts[i]); err != nil {
			return err
		}

		if (i%1000 == 0 || i == len(contracts)-1) && i > 0 {
			logger.Info("Saving updated data from %d to %d...", lastIdx, i)
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

	logger.Info("Done")

	return nil
}
