package transfer

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser struct {
	Storage models.GeneralRepository

	rpc        noderpc.INode
	events     TokenEvents
	stackTrace *stacktrace.StackTrace

	network  string
	chainID  string
	gasLimit int64

	withoutViews bool
}

// NewParser -
func NewParser(rpc noderpc.INode, tzipRepo tzip.Repository, blocks block.Repository, storage models.GeneralRepository, opts ...ParserOption) (*Parser, error) {
	tp := &Parser{
		rpc:     rpc,
		Storage: storage,
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
		parameters := getParameters(operation.Parameters)
		for i := range operation.Tags {
			switch operation.Tags[i] {
			case consts.FA12Tag:
				return p.makeFA12Transfers(operation, parameters)
			case consts.FA2Tag:
				return p.makeFA2Transfers(operation, parameters)
			}
		}
	}
	return nil, nil
}

func (p *Parser) executeEvents(impl tzip.EventImplementation, name string, operation operation.Operation, operationModels []models.Model) ([]*transfer.Transfer, error) {
	if operation.Kind != consts.Transaction {
		return nil, nil
	}

	var event events.Event
	var err error

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
		ctx.Parameters = operation.Parameters
		ctx.Entrypoint = operation.Entrypoint
		event, err = events.NewMichelsonParameter(impl, name)
		if err != nil {
			return nil, err
		}
		return p.makeTransfersFromBalanceEvents(event, ctx, operation, true)
	case impl.MichelsonExtendedStorageEvent.Is(operation.Entrypoint):
		ctx.Parameters = operation.DeffatedStorage
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

func (p *Parser) transferPostprocessing(transfers []*transfer.Transfer, operation operation.Operation) {
	if p.stackTrace.Empty() {
		return
	}
	for i := range transfers {
		p.setParentEntrypoint(operation, transfers[i])
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

func (p *Parser) makeFA12Transfers(operation operation.Operation, parameters gjson.Result) ([]*transfer.Transfer, error) {
	t := transfer.EmptyTransfer(operation)
	fromAddr, err := getAddress(parameters.Get("args.0"))
	if err != nil {
		return nil, err
	}
	toAddr, err := getAddress(parameters.Get("args.1.args.0"))
	if err != nil {
		return nil, err
	}
	t.From = fromAddr
	t.To = toAddr

	if err := t.SetAmountFromString(parameters.Get("args.1.args.1.int").String()); err != nil {
		return nil, fmt.Errorf("makeFA12Transfers error: %s %s %w", operation.Hash, operation.Network, err)
	}

	p.setParentEntrypoint(operation, t)

	return []*transfer.Transfer{t}, nil
}

func (p *Parser) makeFA2Transfers(operation operation.Operation, parameters gjson.Result) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, from := range parameters.Array() {
		fromAddr, err := getAddress(from.Get("args.0"))
		if err != nil {
			return nil, err
		}
		for _, to := range from.Get("args.1").Array() {
			toAddr, err := getAddress(to.Get("args.0"))
			if err != nil {
				return nil, err
			}
			transfer := transfer.EmptyTransfer(operation)
			transfer.From = fromAddr
			transfer.To = toAddr
			if err := transfer.SetAmountFromString(to.Get("args.1.args.1.int").String()); err != nil {
				return nil, fmt.Errorf("makeFA2Transfers error: %s %s %w", operation.Hash, operation.Network, err)
			}
			transfer.TokenID = to.Get("args.1.args.0.int").Int()

			p.setParentEntrypoint(operation, transfer)

			transfers = append(transfers, transfer)
		}
	}
	return transfers, nil
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
