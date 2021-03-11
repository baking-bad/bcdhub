package events

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// MichelsonInitialStorage -
type MichelsonInitialStorage struct {
	Sections

	name   string
	parser tokenbalance.Parser
}

// NewMichelsonInitialStorage -
func NewMichelsonInitialStorage(impl tzip.EventImplementation, name string) (*MichelsonInitialStorage, error) {
	retType, err := ast.NewTypedAstFromBytes(impl.MichelsonInitialStorageEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	parser, err := tokenbalance.GetParser(name, retType)
	if err != nil {
		return nil, err
	}
	return &MichelsonInitialStorage{
		Sections: Sections{
			Parameter:  impl.MichelsonInitialStorageEvent.Parameter,
			Code:       impl.MichelsonInitialStorageEvent.Code,
			ReturnType: impl.MichelsonInitialStorageEvent.ReturnType,
		},

		name:   name,
		parser: parser,
	}, nil
}

// Parse -
func (event *MichelsonInitialStorage) Parse(response noderpc.RunCodeResponse) []tokenbalance.TokenBalance {
	balances, err := event.parser.Parse(response.Storage)
	if err != nil {
		return nil
	}
	return balances
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (event *MichelsonInitialStorage) Normalize(value *ast.TypedAst) []byte {
	b, err := value.ToParameters("")
	if err != nil {
		return nil
	}
	return b
}
