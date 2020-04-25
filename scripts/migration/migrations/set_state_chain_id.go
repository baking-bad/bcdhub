package migrations

import (
	"github.com/baking-bad/bcdhub/internal/logger"
)

// SetStateChainID - migration that set chain id to state model
type SetStateChainID struct{}

// Description -
func (m *SetStateChainID) Description() string {
	return "set chain id to state model"
}

// Do - migrate function
func (m *SetStateChainID) Do(ctx *Context) error {
	for _, network := range []string{"mainnet", "babylonnet", "carthagenet", "zeronet"} { // TODO:
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
		state, err := ctx.ES.CurrentState(network)
		if err != nil {
			return err
		}
		state.ChainID = header.ChainID

		if _, err := ctx.ES.UpdateDoc("state", state.ID, state); err != nil {
			return err
		}
		logger.Info("%s updated chain id", network)
	}

	logger.Success("Done")
	return nil
}
