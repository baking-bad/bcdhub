package transfer

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser struct {
	rpc   noderpc.INode
	es    elastic.IElastic
	views TokenEvents
}

// NewParser -
func NewParser(rpc noderpc.INode, es elastic.IElastic, opts ...ParserOption) *Parser {
	tp := &Parser{
		rpc: rpc,
		es:  es,
	}

	for i := range opts {
		opts[i](tp)
	}

	return tp
}

// Parse -
func (p *Parser) Parse(operation models.Operation) ([]*models.Transfer, error) {
	if view, ok := p.views.GetByOperation(operation); ok {
		return p.runView(view, operation)
	} else if operation.Entrypoint == "transfer" {
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

func (p *Parser) makeFA12Transfers(operation models.Operation, parameters gjson.Result) ([]*models.Transfer, error) {
	transfer := models.EmptyTransfer(operation)
	fromAddr, err := getAddress(parameters.Get("args.0"))
	if err != nil {
		return nil, err
	}
	toAddr, err := getAddress(parameters.Get("args.1.args.0"))
	if err != nil {
		return nil, err
	}
	transfer.From = fromAddr
	transfer.To = toAddr
	transfer.Amount = parameters.Get("args.1.args.1.int").Float()
	return []*models.Transfer{transfer}, nil
}

func (p *Parser) makeFA2Transfers(operation models.Operation, parameters gjson.Result) ([]*models.Transfer, error) {
	transfers := make([]*models.Transfer, 0)
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
			transfer := models.EmptyTransfer(operation)
			transfer.From = fromAddr
			transfer.To = toAddr
			transfer.Amount = to.Get("args.1.args.1.int").Float()
			transfer.TokenID = to.Get("args.1.args.0.int").Int()
			transfers = append(transfers, transfer)
		}
	}
	return transfers, nil
}

func (p Parser) runView(event tzip.EventImplementation, operation models.Operation) ([]*models.Transfer, error) {
	parser, err := event.MichelsonParameterView.GetParser()
	if err != nil {
		return nil, err
	}
	state, err := p.es.GetLastBlock(operation.Network)
	if err != nil {
		return nil, err
	}
	protocol, err := p.es.GetProtocol(operation.Network, "", -1)
	if err != nil {
		return nil, err
	}

	parameter := normalizeParameter(operation.Parameters)
	storage := gjson.Parse(`[]`)
	code, err := event.MichelsonParameterView.CodeJSON()
	if err != nil {
		return nil, err
	}

	response, err := p.rpc.RunCode(code, storage, parameter, state.ChainID, operation.Source, operation.Initiator, operation.Entrypoint, operation.Amount, protocol.Constants.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}
	return p.parseResponse(parser, operation, response)
}

func (p Parser) parseResponse(parser tzip.BalanceViewParser, operation models.Operation, response gjson.Result) ([]*models.Transfer, error) {
	newBalances := parser.Parse(response)
	addresses := make([]elastic.TokenBalance, len(newBalances))
	for i := range newBalances {
		addresses[i].Address = newBalances[i].Address
		addresses[i].TokenID = newBalances[i].TokenID
	}

	oldBalances, err := p.es.GetBalances(operation.Network, operation.Destination, operation.Level, addresses...)
	if err != nil {
		return nil, err
	}

	transfers := make([]*models.Transfer, 0)
	for _, balance := range newBalances {

		transfer := models.EmptyTransfer(operation)
		if oldBalance, ok := oldBalances[elastic.TokenBalance{
			TokenID: balance.TokenID,
			Address: balance.Address,
		}]; ok {
			delta := balance.Value - oldBalance
			if delta > 0 {
				transfer.To = balance.Address
			} else {
				transfer.From = balance.Address
			}
			transfer.Amount = float64(delta)
		} else {
			transfer.Amount = float64(balance.Value)
			transfer.To = balance.Address
		}
		transfer.TokenID = balance.TokenID
		transfer.Nonce = operation.Nonce
		transfer.Counter = operation.Counter
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
