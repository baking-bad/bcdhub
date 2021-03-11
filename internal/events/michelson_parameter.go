package events

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

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
func (event *MichelsonParameter) Normalize(value *ast.TypedAst) []byte {
	if !value.IsSettled() {
		return nil
	}

	result := value.Unwrap()
	if result == nil {
		logger.Warning("MichelsonParameter.Normalize: can't unwrap")
		return nil
	}
	b, err := result.ToParameters()
	if err != nil {
		logger.Warning("MichelsonParameter.Normalize %s", err.Error())
		return nil
	}
	return b
}
