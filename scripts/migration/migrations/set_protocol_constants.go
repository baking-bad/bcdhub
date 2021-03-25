package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
)

// SetProtocolConstants - migration that set constants for protocol
type SetProtocolConstants struct{}

// Key -
func (m *SetProtocolConstants) Key() string {
	return "protocol_constants"
}

// Description -
func (m *SetProtocolConstants) Description() string {
	return "set constants for protocol"
}

// Do - migrate function
func (m *SetProtocolConstants) Do(ctx *config.Context) error {
	protocols, err := ctx.Protocols.GetAll()
	if err != nil {
		return err
	}

	updatedModels := make([]models.Model, 0)
	for i := range protocols {
		if protocols[i].StartLevel == protocols[i].EndLevel && protocols[i].EndLevel == 0 {
			protocols[i].Constants = &protocol.Constants{}
			updatedModels = append(updatedModels, &protocols[i])
			continue
		}

		if protocols[i].Network == "zeronet" {
			continue
		}

		rpc, err := ctx.GetRPC(protocols[i].Network)
		if err != nil {
			return err
		}
		constants, err := rpc.GetNetworkConstants(protocols[i].EndLevel)
		if err != nil {
			return err
		}
		protocols[i].Constants = &protocol.Constants{
			CostPerByte:                  constants.CostPerByte,
			HardGasLimitPerOperation:     constants.HardGasLimitPerOperation,
			HardStorageLimitPerOperation: constants.HardStorageLimitPerOperation,
			TimeBetweenBlocks:            constants.TimeBetweenBlocks[0],
		}

		logger.Info("%##v", protocols[i])
		updatedModels = append(updatedModels, &protocols[i])
	}
	return ctx.Storage.Save(updatedModels)
}
