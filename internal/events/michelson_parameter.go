package events

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/tidwall/gjson"
)

// MichelsonParameter -
type MichelsonParameter struct {
	Sections

	name   string
	parser tokenbalance.Parser
}

// NewMichelsonParameter -
func NewMichelsonParameter(impl tzip.EventImplementation, name string) (*MichelsonParameter, error) {
	parser, err := tokenbalance.GetParser(name, impl.MichelsonParameterEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	return &MichelsonParameter{
		Sections: Sections{
			Parameter:  impl.MichelsonParameterEvent.Parameter,
			Code:       impl.MichelsonParameterEvent.Code,
			ReturnType: impl.MichelsonParameterEvent.ReturnType,
		},

		name:   name,
		parser: parser,
	}, nil
}

// Parse -
func (event *MichelsonParameter) Parse(response gjson.Result) []tokenbalance.TokenBalance {
	balances := make([]tokenbalance.TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balance, err := event.parser.Parse(item)
		if err != nil {
			continue
		}
		balances = append(balances, balance)
	}
	return balances
}

// Normalize - `value` is `Operation.Parameters`
func (event *MichelsonParameter) Normalize(value string) gjson.Result {
	p := gjson.Parse(value)
	if p.Get("value").Exists() {
		p = p.Get("value")
	}

	for prim := p.Get("prim").String(); prim == "Right" || prim == "Left"; prim = p.Get("prim").String() {
		p = p.Get("args.0")
	}
	return p
}
