package transfer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer/trees"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Parser -
type Parser struct {
	tokenBalances tokenbalance.Repository

	rpc        noderpc.INode
	shareDir   string
	stackTrace *stacktrace.StackTrace

	network  modelTypes.Network
	chainID  string
	gasLimit int64

	withoutViews bool
}

var globalEvents *TokenEvents

// NewParser -
func NewParser(rpc noderpc.INode, tzipRepo tzip.Repository, blocks block.Repository, tokenBalances tokenbalance.Repository, shareDir string, opts ...ParserOption) (*Parser, error) {
	tp := &Parser{
		rpc:           rpc,
		tokenBalances: tokenBalances,
		shareDir:      shareDir,
	}

	for i := range opts {
		opts[i](tp)
	}

	if tp.stackTrace == nil {
		tp.stackTrace = stacktrace.New()
	}

	switch {
	case tp.withoutViews && globalEvents == nil:
		globalEvents = EmptyTokenEvents()
	case tp.withoutViews && globalEvents != nil:
	case !tp.withoutViews && globalEvents == nil:
		tokenEvents, err := NewTokenEvents(tzipRepo)
		if err != nil {
			return nil, err
		}
		globalEvents = tokenEvents
	case !tp.withoutViews && globalEvents != nil:
		if err := globalEvents.Update(tzipRepo); err != nil {
			return nil, err
		}
	}

	if tp.network != modelTypes.Empty && tp.chainID == "" {
		state, err := blocks.Last(tp.network)
		if err != nil {
			return nil, err
		}
		tp.chainID = state.ChainID
	}
	return tp, nil
}

// Parse -
func (p *Parser) Parse(diffs []*bigmapdiff.BigMapDiff, protocol string, operation *operation.Operation) error {
	if !operation.IsTransaction() {
		return nil
	}

	if impl, name, ok := globalEvents.GetByOperation(*operation); ok {
		return p.executeEvents(impl, name, protocol, diffs, operation)
	}

	if operation.IsEntrypoint(consts.TransferEntrypoint) {
		switch {
		case operation.Tags.Has(modelTypes.FA2Tag):
			return p.makeFA2Transfers(operation)
		case operation.Tags.Has(modelTypes.FA12Tag):
			return p.makeFA12Transfers(operation)
		}
	}
	return nil
}

func (p *Parser) executeEvents(impl tzip.EventImplementation, name, protocol string, diffs []*bigmapdiff.BigMapDiff, operation *operation.Operation) error {
	if !operation.IsApplied() {
		return nil
	}

	var event events.Event

	ctx := events.Context{
		Network:                  p.network,
		Protocol:                 protocol,
		Source:                   operation.Source,
		Amount:                   operation.Amount,
		Initiator:                operation.Initiator,
		ChainID:                  p.chainID,
		HardGasLimitPerOperation: p.gasLimit,
	}

	switch {
	case impl.MichelsonParameterEvent != nil && impl.MichelsonParameterEvent.Is(operation.Entrypoint):
		parameter, err := operation.AST.ParameterType()
		if err != nil {
			return err
		}
		param := types.NewParameters(operation.Parameters)
		subTree, err := parameter.FromParameters(param)
		if err != nil {
			return err
		}
		ctx.Parameters = subTree
		ctx.Entrypoint = operation.Entrypoint
		event, err = events.NewMichelsonParameter(impl, name)
		if err != nil {
			return err
		}
		return p.makeTransfersFromBalanceEvents(event, ctx, operation, true)
	case impl.MichelsonExtendedStorageEvent != nil && impl.MichelsonExtendedStorageEvent.Is(operation.Entrypoint):
		storage, err := operation.AST.StorageType()
		if err != nil {
			return err
		}
		var deffattedStorage ast.UntypedAST
		if err := json.Unmarshal(operation.DeffatedStorage, &deffattedStorage); err != nil {
			return err
		}
		if err := storage.Settle(deffattedStorage); err != nil {
			return err
		}
		ctx.Parameters = storage
		ctx.Entrypoint = consts.DefaultEntrypoint
		bmd := make([]bigmapdiff.BigMapDiff, 0)
		for i := range diffs {
			if diffs[i].OperationHash == operation.Hash &&
				diffs[i].OperationCounter == operation.Counter &&
				helpers.IsInt64PointersEqual(diffs[i].OperationNonce, operation.Nonce) {
				bmd = append(bmd, *diffs[i])
			}
		}
		event, err = events.NewMichelsonExtendedStorage(impl, name, bmd)
		if err != nil {
			return err
		}
		return p.makeTransfersFromBalanceEvents(event, ctx, operation, false)
	default:
		return nil
	}
}

func (p *Parser) makeTransfersFromBalanceEvents(event events.Event, ctx events.Context, operation *operation.Operation, isDelta bool) error {
	balances, err := events.Execute(p.rpc, event, ctx)
	if err != nil {
		logger.Errorf("Event of %s %s: %s", operation.Network, operation.Destination, err.Error())
		return nil
	}

	parser := NewDefaultBalanceParser(p.tokenBalances)
	if isDelta {
		operation.Transfers, err = parser.Parse(balances, *operation)
	} else {
		operation.Transfers, err = parser.ParseBalances(p.network, operation.Destination, balances, *operation)
	}
	if err != nil {
		return err
	}
	p.transferPostprocessing(operation)

	return err
}

func (p *Parser) transferPostprocessing(operation *operation.Operation) {
	if p.stackTrace.Empty() {
		return
	}
	if len(operation.Transfers) == 0 {
		return
	}
	for i := range operation.Transfers {
		p.setParentEntrypoint(operation, operation.Transfers[i])
	}
}

func (p *Parser) makeFA12Transfers(operation *operation.Operation) error {
	node, err := getNode(operation)
	if err != nil {
		if operation.Status == modelTypes.OperationStatusApplied {
			return err
		}
		return nil
	}
	if node == nil {
		return nil
	}

	operation.Transfers, err = trees.MakeFa1_2Transfers(node, *operation)
	if err != nil {
		return err
	}
	p.transferPostprocessing(operation)
	return nil
}

func (p *Parser) makeFA2Transfers(operation *operation.Operation) error {
	node, err := getNode(operation)
	if err != nil {
		if operation.Status == modelTypes.OperationStatusApplied {
			return err
		}
		return nil
	}
	if node == nil {
		return nil
	}
	operation.Transfers, err = trees.MakeFa2Transfers(node, *operation)
	if err != nil {
		return err
	}
	p.transferPostprocessing(operation)
	return nil
}

func getNode(operation *operation.Operation) (ast.Node, error) {
	param, err := operation.AST.ParameterType()
	if err != nil {
		return nil, err
	}
	params := types.NewParameters(operation.Parameters)

	subTree, err := param.FromParameters(params)
	if err != nil {
		return nil, err
	}

	node, _ := subTree.UnwrapAndGetEntrypointName()

	return node, nil
}

func (p Parser) setParentEntrypoint(operation *operation.Operation, transfer *transfer.Transfer) {
	if p.stackTrace.Empty() {
		return
	}
	item := p.stackTrace.Get(*operation)
	if item == nil || item.ParentID == -1 {
		return
	}
	parent := p.stackTrace.GetByID(item.ParentID)
	if parent == nil {
		return
	}

	transfer.Parent = parent.Entrypoint
}
