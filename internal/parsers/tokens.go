package parsers

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// TokenKey -
type TokenKey struct {
	Address    string
	Network    string
	Entrypoint string
}

// TokenViews -
type TokenViews map[TokenKey]database.TokenViewImplementation

// NewTokenViews -
func NewTokenViews(db database.DB) (TokenViews, error) {
	tokens, err := db.GetTokens()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	views := make(TokenViews)
	for _, token := range tokens {
		if len(token.Metadata.Views) == 0 {
			continue
		}

		for _, view := range token.Metadata.Views {
			for _, implementation := range view.Implementations {
				for _, entrypoint := range implementation.MichelsonParameterView.Entrypoints {
					views[TokenKey{
						Address:    token.Contract,
						Network:    token.Network,
						Entrypoint: entrypoint,
					}] = implementation
				}
			}
		}
	}

	return views, nil
}

// Get -
func (views TokenViews) Get(address, network, entrypoint string) (database.TokenViewImplementation, bool) {
	view, ok := views[TokenKey{
		Address:    address,
		Network:    network,
		Entrypoint: entrypoint,
	}]
	return view, ok
}

// GetByOperation -
func (views TokenViews) GetByOperation(operation models.Operation) (database.TokenViewImplementation, bool) {
	view, ok := views[TokenKey{
		Address:    operation.Destination,
		Network:    operation.Network,
		Entrypoint: operation.Entrypoint,
	}]
	return view, ok
}

// TransferParser -
type TransferParser struct {
	rpc   noderpc.INode
	es    elastic.IElastic
	views TokenViews
}

// NewTransferParser -
func NewTransferParser(rpc noderpc.INode, es elastic.IElastic) TransferParser {
	return TransferParser{
		rpc: rpc,
		es:  es,
	}
}

// SetViews -
func (p *TransferParser) SetViews(views TokenViews) {
	p.views = views
}

// Parse -
func (p TransferParser) Parse(operation models.Operation) ([]*models.Transfer, error) {
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

func (p TransferParser) makeFA12Transfers(operation models.Operation, parameters gjson.Result) ([]*models.Transfer, error) {
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
	transfer.Amount = parameters.Get("args.1.args.1.int").Int()
	return []*models.Transfer{transfer}, nil
}

func (p TransferParser) makeFA2Transfers(operation models.Operation, parameters gjson.Result) ([]*models.Transfer, error) {
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
			transfer.Amount = to.Get("args.1.args.1.int").Int()
			transfer.TokenID = to.Get("args.1.args.0.int").Int()
			transfers = append(transfers, transfer)
		}
	}
	return transfers, nil
}

func (p TransferParser) runView(view database.TokenViewImplementation, operation models.Operation) ([]*models.Transfer, error) {
	parser, err := view.MichelsonParameterView.GetParser()
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
	code, err := view.MichelsonParameterView.CodeJSON()
	if err != nil {
		return nil, err
	}

	response, err := p.rpc.RunCode(code, storage, parameter, state.ChainID, "", "", operation.Entrypoint, 0, protocol.Constants.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}
	return p.parseResponse(parser, operation, response)
}

func (p TransferParser) parseResponse(parser database.BalanceViewParser, operation models.Operation, response gjson.Result) ([]*models.Transfer, error) {
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
			transfer.Amount = delta
		} else {
			transfer.Amount = balance.Value
			transfer.To = balance.Address
		}
		transfer.TokenID = balance.TokenID
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

func normalizeParameter(params string) gjson.Result {
	parameter := gjson.Parse(params)
	if parameter.Get("value").Exists() {
		parameter = parameter.Get("value")
	}

	for prim := parameter.Get("prim").String(); prim == "Right" || prim == "Left"; prim = parameter.Get("prim").String() {
		parameter = parameter.Get("args.0")
	}
	return parameter
}

func getParameters(str string) gjson.Result {
	parameters := gjson.Parse(str)
	if !parameters.Get("value").Exists() {
		return parameters
	}
	parameters = parameters.Get("value")
	for end := false; !end; {
		prim := parameters.Get("prim|@lower").String()
		end = prim != consts.LEFT && prim != consts.RIGHT
		if !end {
			parameters = parameters.Get("args.0")
		}
	}
	return parameters
}

func getAddress(data gjson.Result) (string, error) {
	if data.Get("string").Exists() {
		return data.Get("string").String(), nil
	}

	if data.Get("bytes").Exists() {
		return unpack.Address(data.Get("bytes").String())
	}
	return "", errors.Errorf("Unknown address data: %s", data.Raw)
}
