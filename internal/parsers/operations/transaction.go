package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
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
func (p Transaction) Parse(data noderpc.Operation) ([]models.Model, error) {
	tx := operation.Operation{
		Network:      p.network,
		Hash:         p.hash,
		Protocol:     p.head.Protocol,
		Level:        p.head.Level,
		Timestamp:    p.head.Timestamp,
		Kind:         data.Kind,
		Initiator:    data.Source,
		Source:       data.Source,
		Fee:          data.Fee,
		Counter:      data.Counter,
		GasLimit:     data.GasLimit,
		StorageLimit: data.StorageLimit,
		Amount:       *data.Amount,
		Destination:  *data.Destination,
		Delegate:     data.Delegate,
		Nonce:        data.Nonce,
		Parameters:   data.Parameters,
		ContentIndex: p.contentIdx,
	}

	p.fillInternal(&tx)

	parseOperationResult(data, &tx)

	tx.SetBurned(p.constants)
	txModels := []models.Model{&tx}

	script, err := fetch.Contract(tx.Destination, tx.Network, tx.Protocol, p.shareDir)
	if err != nil {
		return nil, err
	}
	tx.Script = script

	if err := tx.InitScript(); err != nil {
		return nil, err
	}

	if tx.IsApplied() {
		appliedModels, err := p.appliedHandler(data, &tx)
		if err != nil {
			return nil, err
		}
		txModels = append(txModels, appliedModels...)
	}

	if err := p.getEntrypoint(&tx); err != nil {
		return nil, err
	}

	if err := setTags(p.Contracts, p.Storage, &tx); err != nil {
		return nil, err
	}

	p.stackTrace.Add(tx)

	if !tezerrors.HasParametersError(tx.Errors) {
		transfers, err := p.transferParser.Parse(tx, txModels)
		if err != nil {
			if !errors.Is(err, noderpc.InvalidNodeResponse{}) {
				return nil, err
			}
			logger.With(&tx).Warning(err.Error())
		}
		for i := range transfers {
			txModels = append(txModels, transfers[i])
		}

		if !tx.HasTag(consts.LedgerTag) {
			balanceUpdates := transferParsers.UpdateTokenBalances(transfers)
			txModels = append(txModels, balanceUpdates...)
		}
	}
	return txModels, nil
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

func (p Transaction) appliedHandler(item noderpc.Operation, op *operation.Operation) ([]models.Model, error) {
	if !bcd.IsContract(op.Destination) || !op.IsApplied() {
		return nil, nil
	}

	resultModels := make([]models.Model, 0)

	rs, err := p.storageParser.Parse(item, op)
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	op.DeffatedStorage = rs.DeffatedStorage

	resultModels = append(resultModels, rs.Models...)

	migration, err := NewMigration().Parse(item, op)
	if err != nil {
		return nil, err
	}
	if migration != nil {
		resultModels = append(resultModels, migration)
	}

	return resultModels, nil
}

func (p Transaction) getEntrypoint(tx *operation.Operation) error {
	if !tx.IsCall() {
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
