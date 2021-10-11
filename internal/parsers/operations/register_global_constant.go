package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// RegisterGlobalConstant -
type RegisterGlobalConstant struct {
	*ParseParams
}

// NewRegisterGlobalConstant -
func NewRegisterGlobalConstant(params *ParseParams) RegisterGlobalConstant {
	return RegisterGlobalConstant{params}
}

// Parse -
func (p RegisterGlobalConstant) Parse(data noderpc.Operation) (*parsers.Result, error) {
	result := parsers.NewResult()

	proto, err := p.ctx.CachedProtocolByHash(p.network, p.head.Protocol)
	if err != nil {
		return nil, err
	}

	registerGlobalConstant := operation.Operation{
		Network:      p.network,
		Hash:         p.hash,
		ProtocolID:   proto.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         types.NewOperationKind(data.Kind),
		Initiator:    data.Source,
		Source:       data.Source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		StorageLimit: data.StorageLimit,
		Nonce:        data.Nonce,
		ContentIndex: p.contentIdx,
	}
	parseOperationResult(data, &registerGlobalConstant)
	p.stackTrace.Add(registerGlobalConstant)

	result.Operations = append(result.Operations, &registerGlobalConstant)
	if registerGlobalConstant.IsApplied() {
		if err != nil {
			return nil, err
		}
		result.GlobalConstants = append(result.GlobalConstants, NewGlobalConstant().Parse(data, registerGlobalConstant))
	}
	return result, nil
}
