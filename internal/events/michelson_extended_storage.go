package events

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// MichelsonExtendedStorage -
type MichelsonExtendedStorage struct {
	Sections

	name   string
	parser tokenbalance.Parser

	protocol    string
	operationID string
	contract    string
	bmd         []bigmapdiff.BigMapDiff
}

// NewMichelsonExtendedStorage -
func NewMichelsonExtendedStorage(impl tzip.EventImplementation, name, protocol, operationID, contract string, bmd []bigmapdiff.BigMapDiff) (*MichelsonExtendedStorage, error) {
	retType, err := ast.NewTypedAstFromBytes(impl.MichelsonExtendedStorageEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	parser, err := tokenbalance.GetParser(name, retType)
	if err != nil {
		return nil, err
	}
	return &MichelsonExtendedStorage{
		Sections: Sections{
			Parameter:  impl.MichelsonExtendedStorageEvent.Parameter,
			Code:       impl.MichelsonExtendedStorageEvent.Code,
			ReturnType: impl.MichelsonExtendedStorageEvent.ReturnType,
		},

		name:        name,
		parser:      parser,
		protocol:    protocol,
		operationID: operationID,
		bmd:         bmd,
		contract:    contract,
	}, nil
}

// Parse -
func (mes *MichelsonExtendedStorage) Parse(response noderpc.RunCodeResponse) []tokenbalance.TokenBalance {
	balances, err := mes.parser.Parse(response.Storage)
	if err != nil {
		return nil
	}
	return balances
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (mes *MichelsonExtendedStorage) Normalize(value *ast.TypedAst) []byte {
	if !value.IsSettled() {
		return nil
	}

	if err := storage.Enrich(value, mes.bmd, true, false); err != nil {
		logger.Warning("MichelsonExtendedStorage.Normalize %s", err.Error())
		return nil
	}

	b, err := value.ToParameters("")
	if err != nil {
		logger.Warning("MichelsonExtendedStorage.Normalize %s", err.Error())
		return nil
	}
	return b
}
