package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
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
	protocols := make([]models.Protocol, 0)
	if err := ctx.ES.GetAll(&protocols); err != nil {
		return err
	}

	updatedModels := make([]elastic.Model, len(protocols))
	for i := range protocols {
		if protocols[i].StartLevel == protocols[i].EndLevel && protocols[i].EndLevel == 0 {
			protocols[i].Constants = models.Constants{}
			updatedModels[i] = &protocols[i]
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
		protocols[i].Constants = models.Constants{
			CostPerByte:                  constants.Get("cost_per_byte").Int(),
			HardGasLimitPerOperation:     constants.Get("hard_gas_limit_per_operation").Int(),
			HardStorageLimitPerOperation: constants.Get("hard_storage_limit_per_operation").Int(),
			TimeBetweenBlocks:            constants.Get("time_between_blocks.0").Int(),
		}
		updatedModels[i] = &protocols[i]
	}
	return ctx.ES.BulkUpdate(updatedModels)
}
