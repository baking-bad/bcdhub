package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelsTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/ledger"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/pkg/errors"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Transaction -
type Transaction struct {
	*ParseParams
}

// NewTransaction -
func NewTransaction(params *ParseParams) Transaction {
	return Transaction{params}
}

// Parse -
func (p Transaction) Parse(data noderpc.Operation, result *parsers.Result) error {
	proto, err := p.ctx.Cache.ProtocolByHash(p.network, p.head.Protocol)
	if err != nil {
		return err
	}

	source := account.Account{
		Network: p.network,
		Address: data.Source,
		Type:    modelsTypes.NewAccountType(data.Source),
	}

	tx := operation.Operation{
		Network:      p.network,
		Hash:         p.hash,
		ProtocolID:   proto.ID,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         modelsTypes.NewOperationKind(data.Kind),
		Initiator:    source,
		Source:       source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		StorageLimit: data.StorageLimit,
		Amount:       *data.Amount,
		Destination: account.Account{
			Network: p.network,
			Address: *data.Destination,
			Type:    modelsTypes.NewAccountType(*data.Destination),
		},
		Delegate: account.Account{
			Network: p.network,
			Address: data.Delegate,
			Type:    modelsTypes.NewAccountType(data.Delegate),
		},
		Nonce:        data.Nonce,
		Parameters:   data.Parameters,
		ContentIndex: p.contentIdx,
	}

	p.fillInternal(&tx)

	parseOperationResult(data, &tx)

	tx.SetBurned(p.constants)

	result.Operations = append(result.Operations, &tx)

	if tx.Destination.Type != modelsTypes.AccountTypeContract {
		return nil
	}

	for i := range tx.Errors {
		if tx.Errors[i].Is("contract.non_existing_contract") {
			return nil
		}
	}

	script, err := p.ctx.Cache.ScriptBytes(tx.Network, tx.Destination.Address, proto.SymLink)
	if err != nil {
		if !tx.Internal {
			return nil
		}

		for i := range result.Contracts {
			if tx.Destination.Address == result.Contracts[i].Account.Address {
				switch proto.SymLink {
				case bcd.SymLinkAlpha:
					script, err = result.Contracts[i].Alpha.Full()
					if err != nil {
						return err
					}
				case bcd.SymLinkBabylon:
					script, err = result.Contracts[i].Babylon.Full()
					if err != nil {
						return err
					}
				default:
					return errors.Errorf("unknwon protocol symbolic link: %s", proto.SymLink)
				}
			}
		}
		if script == nil {
			return err
		}
	}
	tx.Script = script

	tx.AST, err = ast.NewScriptWithoutCode(script)
	if err != nil {
		return err
	}

	if err := setTags(p.ctx, nil, &tx); err != nil {
		return err
	}

	if err := p.getEntrypoint(&tx); err != nil {
		return err
	}
	p.stackTrace.Add(tx)

	if tx.IsApplied() {
		if err := p.appliedHandler(data, &tx, result); err != nil {
			return err
		}
	}

	if !tezerrors.HasParametersError(tx.Errors) {
		if err := p.transferParser.Parse(tx.BigMapDiffs, p.head.Protocol, &tx); err != nil {
			if !errors.Is(err, noderpc.InvalidNodeResponse{}) {
				return err
			}
			logger.Warning().Err(err).Msg("transferParser.Parse")
		}
		result.TokenBalances = append(result.TokenBalances, transferParsers.UpdateTokenBalances(tx.Transfers)...)
	}

	if tx.IsApplied() {
		ledgerResult, err := ledger.New(p.ctx.TokenBalances, p.ctx.Accounts).Parse(&tx, p.stackTrace)
		if err != nil {
			return err
		}
		if ledgerResult != nil {
			result.TokenBalances = append(result.TokenBalances, ledgerResult.TokenBalances...)
		}
	}

	return nil
}

func (p Transaction) fillInternal(tx *operation.Operation) {
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

func (p Transaction) appliedHandler(item noderpc.Operation, tx *operation.Operation, result *parsers.Result) error {
	storageResult, err := p.storageParser.Parse(item, tx)
	if err != nil {
		return err
	}
	if storageResult != nil {
		result.Merge(storageResult)
	}

	return NewMigration(p.ctx.Contracts).Parse(item, tx, result)
}

func (p Transaction) getEntrypoint(tx *operation.Operation) error {
	if len(tx.Parameters) == 0 {
		return tx.Entrypoint.Scan(consts.DefaultEntrypoint)
	}

	params := types.NewParameters(tx.Parameters)
	if err := tx.Entrypoint.Scan(params.Entrypoint); err != nil {
		return err
	}

	if !tx.IsApplied() {
		return nil
	}

	param, err := tx.AST.ParameterType()
	if err != nil {
		return err
	}

	subTree, err := param.FromParameters(params)
	if err != nil {
		return err
	}

	node, entrypointName := subTree.UnwrapAndGetEntrypointName()
	if node == nil {
		return nil
	}

	return tx.Entrypoint.Scan(entrypointName)
}
