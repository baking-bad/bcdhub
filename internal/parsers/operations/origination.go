package operations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Origination -
type Origination struct {
	*ParseParams
}

// NewOrigination -
func NewOrigination(params *ParseParams) Origination {
	return Origination{params}
}

var delegatorContract = []byte(`{"code":[{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"lambda","args":[{"prim":"unit"},{"prim":"list","args":[{"prim":"operation"}]}],"annots":["%do"]},{"prim":"unit","annots":["%default"]}]}]},{"prim":"storage","args":[{"prim":"key_hash"}]},{"prim":"code","args":[[[[{"prim":"DUP"},{"prim":"CAR"},{"prim":"DIP","args":[[{"prim":"CDR"}]]}]],{"prim":"IF_LEFT","args":[[{"prim":"PUSH","args":[{"prim":"mutez"},{"int":"0"}]},{"prim":"AMOUNT"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],[{"prim":"DIP","args":[[{"prim":"DUP"}]]},{"prim":"SWAP"}],{"prim":"IMPLICIT_ACCOUNT"},{"prim":"ADDRESS"},{"prim":"SENDER"},[[{"prim":"COMPARE"},{"prim":"EQ"}],{"prim":"IF","args":[[],[[{"prim":"UNIT"},{"prim":"FAILWITH"}]]]}],{"prim":"UNIT"},{"prim":"EXEC"},{"prim":"PAIR"}],[{"prim":"DROP"},{"prim":"NIL","args":[{"prim":"operation"}]},{"prim":"PAIR"}]]}]]}],"storage":{"bytes":"0079943a60100e0394ac1c8f6ccfaeee71ec9c2d94"}}`)

// Parse -
func (p Origination) Parse(ctx context.Context, data noderpc.Operation, store parsers.Store) error {
	source := account.Account{
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
		Level:   p.head.Level,
	}

	origination := operation.Operation{
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
		Amount:       *data.Balance,
		Delegate: account.Account{
			Address: data.Delegate,
			Type:    types.NewAccountType(data.Delegate),
			Level:   p.head.Level,
		},
		Parameters:   data.Parameters,
		Nonce:        data.Nonce,
		ContentIndex: p.contentIdx,
		Script:       data.Script,
	}

	if origination.Script == nil {
		origination.Script = delegatorContract
	}

	p.fillInternal(&origination)

	parseOperationResult(data, &origination)

	origination.SetBurned(*p.protocol.Constants)

	p.stackTrace.Add(origination)

	if origination.IsApplied() {
		if err := p.appliedHandler(ctx, data, &origination, store); err != nil {
			return err
		}
	}

	store.AddOperations(&origination)

	return nil
}

func (p Origination) appliedHandler(ctx context.Context, item noderpc.Operation, origination *operation.Operation, store parsers.Store) error {
	if origination == nil || store == nil {
		return nil
	}

	if p.specific.NeedReceiveRawStorage {
		rawStorage, err := p.ctx.RPC.GetScriptStorageRaw(ctx, origination.Destination.Address, origination.Level)
		if err != nil {
			return err
		}
		origination.DeffatedStorage = rawStorage
	}

	if err := p.specific.ContractParser.Parse(ctx, origination, store); err != nil {
		return err
	}

	contracts := store.ListContracts()
	if err := setTags(ctx, p.ctx, contracts[0], origination); err != nil {
		return err
	}

	if err := p.specific.StorageParser.ParseOrigination(ctx, item, origination, store); err != nil {
		return err
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
