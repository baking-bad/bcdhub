package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// SetStateChainID - migration that set chain id to state model
type SetStateChainID struct{}

// Key -
func (m *SetStateChainID) Key() string {
	return "state_chain_id"
}

// Description -
func (m *SetStateChainID) Description() string {
	return "set chain id to state model"
}

// Do - migrate function
func (m *SetStateChainID) Do(ctx *config.Context) error {
	for _, network := range ctx.Config.Migrations.Networks {
		logger.Info("Getting chain id for %s", network)
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return err
		}
		header, err := rpc.GetHead()
		if err != nil {
			return err
		}
		logger.Info("Chain ID: %s", header.ChainID)
		state, err := ctx.ES.GetLastBlock(network)
		if err != nil {
			return err
		}
		state.ChainID = header.ChainID

		if _, err := ctx.ES.UpdateDoc(&state); err != nil {
			return err
		}
		logger.Info("%s updated chain id", network)
	}

	logger.Success("Done")
	return nil
}
