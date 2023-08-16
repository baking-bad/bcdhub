package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// SrOriginate -
type SrOriginate struct {
	*ParseParams
}

// NewSrOriginate -
func NewSrOriginate(params *ParseParams) SrOriginate {
	return SrOriginate{params}
}

// Parse -
func (p SrOriginate) Parse(data noderpc.Operation, store parsers.Store) error {
	source := account.Account{
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
	}

	operation := operation.Operation{
		Source:       source,
		Initiator:    source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		StorageLimit: data.StorageLimit,
		GasLimit:     data.GasLimit,
		Hash:         p.hash,
		ProtocolID:   p.protocol.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         types.NewOperationKind(data.Kind),
		ContentIndex: p.contentIdx,
	}

	p.fillInternal(&operation)
	operation.SetBurned(*p.protocol.Constants)
	parseOperationResult(data, &operation)
	p.stackTrace.Add(operation)

	store.AddOperations(&operation)

	if operation.IsApplied() {
		smartRollup, err := NewSmartRolupParser().Parse(data, operation)
		if err != nil {
			return err
		}
		store.AddSmartRollups(&smartRollup)
		operation.Destination = smartRollup.Address
	}

	return nil
}

func (p SrOriginate) fillInternal(tx *operation.Operation) {
	if p.main == nil {
		p.main = tx
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}
