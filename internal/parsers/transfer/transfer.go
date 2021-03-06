package transfer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer/trees"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Parser -
type Parser struct {
	Storage models.GeneralRepository

	rpc        noderpc.INode
	shareDir   string
	events     TokenEvents
	stackTrace *stacktrace.StackTrace

	network  string
	chainID  string
	gasLimit int64

	withoutViews bool
}

// NewParser -
func NewParser(rpc noderpc.INode, tzipRepo tzip.Repository, blocks block.Repository, storage models.GeneralRepository, shareDir string, opts ...ParserOption) (*Parser, error) {
	tp := &Parser{
		rpc:      rpc,
		Storage:  storage,
		shareDir: shareDir,
	}

	for i := range opts {
		opts[i](tp)
	}

	if tp.stackTrace == nil {
		tp.stackTrace = stacktrace.New()
	}

	if !tp.withoutViews {
		tokenEvents, err := NewTokenEvents(tzipRepo, storage)
		if err != nil {
			return nil, err
		}
		tp.events = tokenEvents
	} else {
		tp.events = make(TokenEvents)
	}

	if tp.network != "" && tp.chainID == "" {
		state, err := blocks.Last(tp.network)
		if err != nil {
			return nil, err
		}
		tp.chainID = state.ChainID
	}
	return tp, nil
}

// Parse -
func (p *Parser) Parse(operation operation.Operation, operationModels []models.Model) ([]*transfer.Transfer, error) {
	if impl, name, ok := p.events.GetByOperation(operation); ok {
		return p.executeEvents(impl, name, operation, operationModels)
	} else if operation.Entrypoint == consts.TransferEntrypoint {
		for i := range operation.Tags {
			switch operation.Tags[i] {
			case consts.FA12Tag:
				return p.makeFA12Transfers(operation)
			case consts.FA2Tag:
				return p.makeFA2Transfers(operation)
			}
		}
	}
	return nil, nil
}

func (p *Parser) executeEvents(impl tzip.EventImplementation, name string, operation operation.Operation, operationModels []models.Model) ([]*transfer.Transfer, error) {
	if operation.Kind != consts.Transaction || !operation.IsApplied() {
		return nil, nil
	}

	var event events.Event

	ctx := events.Context{
		Network:                  p.network,
		Protocol:                 operation.Protocol,
		Source:                   operation.Source,
		Amount:                   operation.Amount,
		Initiator:                operation.Initiator,
		ChainID:                  p.chainID,
		HardGasLimitPerOperation: p.gasLimit,
	}

	switch {
	case impl.MichelsonParameterEvent.Is(operation.Entrypoint):
		data, err := fetch.Contract(operation.Destination, operation.Network, operation.Protocol, p.shareDir)
		if err != nil {
			return nil, err
		}
		script, err := ast.NewScript(data)
		if err != nil {
			return nil, err
		}
		parameter, err := script.ParameterType()
		if err != nil {
			return nil, err
		}
		param := types.NewParameters([]byte(operation.Parameters))
		subTree, err := parameter.FromParameters(param)
		if err != nil {
			return nil, err
		}
		ctx.Parameters = subTree
		ctx.Entrypoint = operation.Entrypoint
		event, err = events.NewMichelsonParameter(impl, name)
		if err != nil {
			return nil, err
		}
		return p.makeTransfersFromBalanceEvents(event, ctx, operation, true)
	case impl.MichelsonExtendedStorageEvent.Is(operation.Entrypoint):
		data, err := fetch.Contract(operation.Destination, operation.Network, operation.Protocol, p.shareDir)
		if err != nil {
			return nil, err
		}
		script, err := ast.NewScript(data)
		if err != nil {
			return nil, err
		}
		storage, err := script.StorageType()
		if err != nil {
			return nil, err
		}
		var deffattedStorage ast.UntypedAST
		if err := json.UnmarshalFromString(operation.DeffatedStorage, &deffattedStorage); err != nil {
			return nil, err
		}
		if err := storage.Settle(deffattedStorage); err != nil {
			return nil, err
		}
		ctx.Parameters = storage
		ctx.Entrypoint = consts.DefaultEntrypoint
		bmd := make([]bigmapdiff.BigMapDiff, 0)
		for i := range operationModels {
			if model, ok := operationModels[i].(*bigmapdiff.BigMapDiff); ok && model.OperationID == operation.ID {
				bmd = append(bmd, *model)
			}
		}
		event, err = events.NewMichelsonExtendedStorage(impl, name, operation.Protocol, operation.GetID(), operation.Destination, bmd)
		if err != nil {
			return nil, err
		}
		return p.makeTransfersFromBalanceEvents(event, ctx, operation, false)
	default:
		return nil, nil
	}
}

func (p *Parser) makeTransfersFromBalanceEvents(event events.Event, ctx events.Context, operation operation.Operation, isDelta bool) ([]*transfer.Transfer, error) {
	balances, err := events.Execute(p.rpc, event, ctx)
	if err != nil {
		return nil, err
	}

	var transfers []*transfer.Transfer

	parser := NewDefaultBalanceParser(p.Storage)
	if isDelta {
		transfers, err = parser.Parse(balances, operation)
	} else {
		transfers, err = parser.ParseBalances(p.network, operation.Destination, balances, operation)
	}
	if err != nil {
		return nil, err
	}
	p.transferPostprocessing(transfers, operation)

	return transfers, err
}

func (p *Parser) transferPostprocessing(transfers []*transfer.Transfer, operation operation.Operation) {
	if p.stackTrace.Empty() {
		return
	}
	for i := range transfers {
		p.setParentEntrypoint(operation, transfers[i])
	}
}

func (p *Parser) makeFA12Transfers(operation operation.Operation) ([]*transfer.Transfer, error) {
	node, err := getNode(operation)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	transfers, err := trees.MakeFa1_2Transfers(node, operation)
	if err != nil {
		return nil, err
	}
	for i := range transfers {
		p.setParentEntrypoint(operation, transfers[i])
	}
	return transfers, nil
}

func (p *Parser) makeFA2Transfers(operation operation.Operation) ([]*transfer.Transfer, error) {
	node, err := getNode(operation)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	transfers, err := trees.MakeFa2Transfers(node, operation)
	if err != nil {
		return nil, err
	}
	for i := range transfers {
		p.setParentEntrypoint(operation, transfers[i])
	}
	return transfers, nil
}

func getNode(operation operation.Operation) (ast.Node, error) {
	var s ast.Script
	if err := json.Unmarshal(operation.Script, &s); err != nil {
		return nil, err
	}

	param, err := s.ParameterType()
	if err != nil {
		return nil, err
	}
	params := types.NewParameters([]byte(operation.Parameters))

	subTree, err := param.FromParameters(params)
	if err != nil {
		return nil, err
	}

	return subTree.Unwrap(), nil
}

func (p Parser) setParentEntrypoint(operation operation.Operation, transfer *transfer.Transfer) {
	if p.stackTrace.Empty() {
		return
	}
	item := p.stackTrace.Get(operation)
	if item == nil || item.ParentID == -1 {
		return
	}
	parent := p.stackTrace.GetByID(item.ParentID)
	if parent == nil {
		return
	}

	transfer.Parent = parent.Entrypoint
}
