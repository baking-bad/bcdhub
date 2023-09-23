package operations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
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
func (p RegisterGlobalConstant) Parse(ctx context.Context, data noderpc.Operation, store parsers.Store) error {
	source := account.Account{
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
	}

	registerGlobalConstant := operation.Operation{
		Hash:         p.hash,
		ProtocolID:   p.protocol.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         types.NewOperationKind(data.Kind),
		Initiator:    source,
		Source:       source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		StorageLimit: data.StorageLimit,
		Nonce:        data.Nonce,
		ContentIndex: p.contentIdx,
	}
	parseOperationResult(data, &registerGlobalConstant)
	p.stackTrace.Add(registerGlobalConstant)

	store.AddOperations(&registerGlobalConstant)
	if registerGlobalConstant.IsApplied() {
		store.AddGlobalConstants(NewGlobalConstant().Parse(data, registerGlobalConstant))
	}
	return nil
}
