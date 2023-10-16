package operations

import (
	"encoding/hex"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// SmartRolupParser -
type SmartRolupParser struct{}

// NewSmartRolupParser -
func NewSmartRolupParser() SmartRolupParser {
	return SmartRolupParser{}
}

// Parse -
func (sr SmartRolupParser) Parse(data noderpc.Operation, operation operation.Operation) (smartrollup.SmartRollup, error) {
	result := data.GetResult()
	rollup := smartrollup.SmartRollup{
		Address: account.Account{
			Address:         result.Address,
			Type:            types.NewAccountType(result.Address),
			Level:           operation.Level,
			OperationsCount: 1,
			LastAction:      operation.Timestamp,
		},
		GenesisCommitmentHash: result.GenesisCommitmentHash,
		PvmKind:               data.PvmKind,
		Type:                  data.ParameterType,
		Level:                 operation.Level,
		Timestamp:             operation.Timestamp,
	}
	if data.Kernel != "" {
		kernel, err := hex.DecodeString(data.Kernel)
		if err != nil {
			return rollup, err
		}
		rollup.Kernel = kernel
	}
	if result.Size != "" {
		size, err := strconv.ParseUint(result.Size, 10, 64)
		if err != nil {
			return rollup, err
		}
		rollup.Size = size
	}
	return rollup, nil
}
