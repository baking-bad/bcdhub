package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/ledger"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
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

	p.fillInternal(&origination)

	parseOperationResult(data, &origination)

	origination.SetBurned(p.constants)

	p.stackTrace.Add(origination)

	if origination.IsApplied() {
		if err := p.appliedHandler(data, &origination, result); err != nil {
			return nil, err
		}
	}

	result.Operations = append(result.Operations, &origination)

	return result, nil
}

func (p Origination) appliedHandler(item noderpc.Operation, origination *operation.Operation, result *parsers.Result) error {
	if origination == nil || result == nil {
		return nil
	}
	if !bcd.IsContract(origination.Destination) || !origination.IsApplied() {
		return nil
	}

	contractResult, err := p.contractParser.Parse(origination)
	if err != nil {
		return err
	}
	result.Contracts = append(result.Contracts, contractResult.Contracts...)

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

	ledgerResult, err := ledger.New(p.ctx.TokenBalances).Parse(origination, p.stackTrace)
	if err != nil {
		return err
	}
	if ledgerResult != nil {
		result.TokenBalances = append(result.TokenBalances, ledgerResult.TokenBalances...)
	}

	if err := p.executeInitialStorageEvent(item.Script, origination, result); err != nil {
		if !errors.Is(err, tokens.ErrNoMetadataKeyInStorage) {
			logger.Err(err)
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
	tzip, err := p.ctx.CachedContractMetadata(origination.Network, origination.Destination)
	if err != nil {
		if p.ctx.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}
	if tzip == nil || len(tzip.Events) == 0 {
		return nil
	}

	for i := range tzip.Events {
		for j := range tzip.Events[i].Implementations {
			impl := tzip.Events[i].Implementations[j]
			if impl.MichelsonInitialStorageEvent == nil || impl.MichelsonInitialStorageEvent.Empty() {
				continue
			}

			event, err := events.NewMichelsonInitialStorage(impl, tzip.Events[i].Name)
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
				Source:                   origination.Source,
				Initiator:                origination.Initiator,
				Amount:                   origination.Amount,
				HardGasLimitPerOperation: p.constants.HardGasLimitPerOperation,
				ChainID:                  p.head.ChainID,
			})
			if err != nil {
				return err
			}

			res, err := transfer.NewDefaultBalanceParser(p.ctx.TokenBalances).Parse(balances, *origination)
			if err != nil {
				return err
			}

			for i := range res {
				origination.Transfers = append(origination.Transfers, res[i])
			}

			for i := range balances {
				result.TokenBalances = append(result.TokenBalances, &tbModel.TokenBalance{
					Network:  tzip.Network,
					Address:  balances[i].Address,
					TokenID:  balances[i].TokenID,
					Contract: tzip.Address,
					Balance:  balances[i].Value,
				})

			}
		}
	}

	return nil
}
