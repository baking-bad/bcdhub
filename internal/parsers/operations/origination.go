package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/ledger"
)

// Origination -
type Origination struct {
	*ParseParams
}

// NewOrigination -
func NewOrigination(params *ParseParams) Origination {
	return Origination{params}
}

// Parse -
func (p Origination) Parse(data noderpc.Operation) (*parsers.Result, error) {
	result := parsers.NewResult()

	proto, err := p.ctx.CachedProtocolByHash(p.network, p.head.Protocol)
	if err != nil {
		return nil, err
	}

	origination := operation.Operation{
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
		Amount:       *data.Balance,
		Delegate:     data.Delegate,
		Parameters:   data.Parameters,
		Nonce:        data.Nonce,
		ContentIndex: p.contentIdx,
		Script:       data.Script,
	}

	result.Operations = append(result.Operations, &origination)

	p.fillInternal(&origination)

	parseOperationResult(data, &origination)

	origination.SetBurned(p.constants)

	p.stackTrace.Add(origination)

	if origination.IsApplied() {
		if err := p.appliedHandler(data, &origination, result); err != nil {
			return nil, err
		}

		if err := setTags(p.ctx, result.Contracts[0], &origination); err != nil {
			return nil, err
		}

		ledgerResult, err := ledger.New(p.ctx.TokenBalances).Parse(&origination, p.stackTrace)
		if err != nil {
			return nil, err
		}
		if ledgerResult != nil {
			result.TokenBalances = append(result.TokenBalances, ledgerResult.TokenBalances...)
		}
	}

	return result, nil
}

func (p Origination) appliedHandler(item noderpc.Operation, origination *operation.Operation, result *parsers.Result) error {
	if !bcd.IsContract(origination.Destination) || !origination.IsApplied() {
		return nil
	}

	contractResult, err := p.contractParser.Parse(origination)
	if err != nil {
		return err
	}
	result.Contracts = append(result.Contracts, contractResult.Contracts...)

	storageResult, err := p.storageParser.Parse(item, origination)
	if err != nil {
		return err
	}
	if storageResult != nil {
		result.Merge(storageResult)
	}

	return nil
}

func (p Origination) fillInternal(tx *operation.Operation) {
	if p.main == nil {
		return
	}

	tx.Counter = p.main.Counter
	tx.Hash = p.main.Hash
	tx.Level = p.main.Level
	tx.Timestamp = p.main.Timestamp
	tx.Internal = true
	tx.Initiator = p.main.Source
}
