package events

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// MichelsonParameter -
type MichelsonParameter struct {
	Sections

	name   string
	parser tokenbalance.Parser
}

// NewMichelsonParameter -
func NewMichelsonParameter(impl tzip.EventImplementation, name string) (*MichelsonParameter, error) {
	retType, err := ast.NewTypedAstFromBytes(impl.MichelsonParameterEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	parser, err := tokenbalance.GetParser(name, retType)
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
func (event *MichelsonParameter) Parse(response noderpc.RunCodeResponse) []tokenbalance.TokenBalance {
	balances, err := event.parser.Parse(response.Storage)
	if err != nil {
		return nil
	}
	return balances
}

// Normalize - `value` is `Operation.Parameters`
func (event *MichelsonParameter) Normalize(value string) []byte {
	params := types.NewParameters([]byte(value))

	var data ast.UntypedAST
	if err := json.UnmarshalFromString(string(params.Value), &data); err != nil {
		logger.Error(err)
		return []byte(value)
	}

	if len(data) == 0 {
		return []byte(value)
	}

	for prim := data[0].Prim; prim == "Right" || prim == "Left"; prim = data[0].Prim {
		data = data[0].Args
	}
	if len(data) == 0 {
		return []byte(value)
	}
	b, err := json.Marshal(data[0])
	if err != nil {
		logger.Error(err)
		return []byte(value)
	}
	return b
}
