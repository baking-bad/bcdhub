package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelsTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
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
func (p Transaction) Parse(data noderpc.Operation) (*parsers.Result, error) {
	result := parsers.NewResult()

	proto, err := p.ctx.CachedProtocolByHash(p.network, p.head.Protocol)
	if err != nil {
		return nil, err
	}

	tx := operation.Operation{
		Network:       p.network,
		Hash:          p.hash,
		ProtocolID:    proto.ID,
		Level:         p.head.Level,
		Timestamp:     p.head.Timestamp,
		Kind:          modelsTypes.NewOperationKind(data.Kind),
		Initiator:     data.Source,
		Source:        data.Source,
		Fee:           data.Fee,
		Counter:       data.Counter,
		GasLimit:      data.GasLimit,
		StorageLimit:  data.StorageLimit,
		Amount:        *data.Amount,
		Destination:   *data.Destination,
		Delegate:      data.Delegate,
		Nonce:         data.Nonce,
		Parameters:    data.Parameters,
		ContentIndex:  p.contentIdx,
		SourceAlias:   p.ctx.CachedAlias(p.network, data.Source),
		DelegateAlias: p.ctx.CachedAlias(p.network, data.Delegate),
	}
	if data.Destination != nil {
		tx.DestinationAlias = p.ctx.CachedAlias(p.network, *data.Destination)
	}

	p.fillInternal(&tx)

	parseOperationResult(data, &tx)

	tx.SetBurned(p.constants)

	result.Operations = append(result.Operations, &tx)

	script, err := p.ctx.CachedScriptBytes(tx.Network, tx.Destination, proto.SymLink)
	if err != nil {
		return nil, err
	}
	tx.Script = script

	tx.AST, err = p.ctx.CachedScript(tx.Network, tx.Destination, proto.SymLink)
	if err != nil {
		return nil, err
	}

	if tx.IsApplied() {
		if err := p.appliedHandler(data, &tx, result); err != nil {
			return nil, err
		}
	}

	if err := p.getEntrypoint(&tx); err != nil {
		return nil, err
	}

	if err := setTags(p.ctx, &tx); err != nil {
		return nil, err
	}

	p.stackTrace.Add(tx)

	if !tezerrors.HasParametersError(tx.Errors) {
		if err := p.transferParser.Parse(result.BigMapDiffs, p.head.Protocol, &tx); err != nil {
			if !errors.Is(err, noderpc.InvalidNodeResponse{}) {
				return nil, err
			}
			logger.With(&tx).Warning(err.Error())
		}
		result.TokenBalances = append(result.TokenBalances, transferParsers.UpdateTokenBalances(tx.Transfers)...)
	}
	return result, nil
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

func (p Transaction) appliedHandler(item noderpc.Operation, op *operation.Operation, result *parsers.Result) error {
	if !bcd.IsContract(op.Destination) || !op.IsApplied() {
		return nil
	}

	rs, err := p.storageParser.Parse(item, op)
	if err != nil {
		return err
	}
	if rs.Empty {
		return nil
	}
	op.DeffatedStorage = rs.DeffatedStorage

	result.Merge(rs.Result)

	migration, err := NewMigration().Parse(item, op)
	if err != nil {
		return err
	}
	if migration != nil {
		result.Migrations = append(result.Migrations, migration)
	}

	return nil
}

func (p Transaction) getEntrypoint(tx *operation.Operation) error {
	if !bcd.IsContract(tx.Destination) {
		return nil
	}

	if len(tx.Parameters) == 0 {
		tx.Entrypoint = consts.DefaultEntrypoint
		return nil
	}

	params := types.NewParameters(tx.Parameters)
	tx.Entrypoint = params.Entrypoint

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
	tx.Entrypoint = entrypointName

	return nil
}
