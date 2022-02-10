package operations

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/tokens"
	"github.com/baking-bad/bcdhub/internal/parsers/ledger"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/pkg/errors"
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
func (p Origination) Parse(data noderpc.Operation, result *parsers.Result) error {
	source := account.Account{
		Network: p.network,
		Address: data.Source,
		Type:    types.NewAccountType(data.Source),
	}

	origination := operation.Operation{
		Network:      p.network,
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
			Network: p.network,
			Address: data.Delegate,
			Type:    types.NewAccountType(data.Delegate),
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
		if err := p.appliedHandler(data, &origination, result); err != nil {
			return err
		}
	}

	result.Operations = append(result.Operations, &origination)

	return nil
}

func (p Origination) appliedHandler(item noderpc.Operation, origination *operation.Operation, result *parsers.Result) error {
	if origination == nil || result == nil {
		return nil
	}

	if err := p.contractParser.Parse(origination, p.protocol.SymLink, result); err != nil {
		return err
	}

	if err := setTags(p.ctx, result.Contracts[0], origination); err != nil {
		return err
	}

	storageResult, err := p.storageParser.Parse(item, origination)
	if err != nil {
		return err
	}
	if storageResult != nil {
		result.Merge(storageResult)
	}

	ledgerResult, err := ledger.New(p.ctx.TokenBalances, p.ctx.Accounts).Parse(origination, p.stackTrace)
	if err != nil {
		return err
	}
	if ledgerResult != nil {
		result.TokenBalances = append(result.TokenBalances, ledgerResult.TokenBalances...)
	}

	if origination.Network == types.Mainnet {
		if err := p.executeInitialStorageEvent(item.Script, origination, result); err != nil {
			if !errors.Is(err, tokens.ErrNoMetadataKeyInStorage) {
				logger.Err(err)
			}
		}
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

func (p Origination) executeInitialStorageEvent(raw []byte, origination *operation.Operation, result *parsers.Result) error {
	if origination == nil || result == nil || origination.Tags.Has(types.LedgerTag) {
		return nil
	}

	contractEvents, err := p.ctx.Cache.Events(origination.Network, origination.Destination.Address)
	if err != nil {
		if p.ctx.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if len(contractEvents) == 0 {
		return nil
	}

	for i := range contractEvents {
		for j := range contractEvents[i].Implementations {
			impl := contractEvents[i].Implementations[j]
			if impl.MichelsonInitialStorageEvent == nil || impl.MichelsonInitialStorageEvent.Empty() {
				continue
			}

			event, err := events.NewMichelsonInitialStorage(impl, contractEvents[i].Name)
			if err != nil {
				return err
			}

			var script noderpc.Script
			if err := json.Unmarshal(raw, &script); err != nil {
				return err
			}

			storageType, err := script.Code.StorageType()
			if err != nil {
				return err
			}

			if err := storageType.SettleFromBytes(script.Storage); err != nil {
				return err
			}

			balances, err := events.Execute(p.rpc, event, events.Context{
				Network:                  origination.Network,
				Parameters:               storageType,
				Source:                   origination.Source.Address,
				Initiator:                origination.Initiator.Address,
				Amount:                   origination.Amount,
				HardGasLimitPerOperation: p.protocol.Constants.HardGasLimitPerOperation,
				ChainID:                  p.head.ChainID,
			})
			if err != nil {
				return err
			}

			res, err := transfer.NewDefaultBalanceParser(p.ctx.TokenBalances, p.ctx.Accounts).Parse(balances, *origination)
			if err != nil {
				return err
			}

			for i := range res {
				origination.Transfers = append(origination.Transfers, res[i])
			}

			for i := range balances {
				result.TokenBalances = append(result.TokenBalances, &tbModel.TokenBalance{
					Network: origination.Network,
					Account: account.Account{
						Address: balances[i].Address,
						Network: origination.Network,
						Type:    types.NewAccountType(balances[i].Address),
					},
					TokenID:  balances[i].TokenID,
					Contract: origination.Destination.Address,
					Balance:  balances[i].Value,
				})

			}
		}
	}

	return nil
}
