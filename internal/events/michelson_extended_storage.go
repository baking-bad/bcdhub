package events

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/tidwall/gjson"
)

// MichelsonExtendedStorage -
type MichelsonExtendedStorage struct {
	Sections

	name   string
	parser tokenbalance.Parser

	protocol    string
	operationID string
	contract    string
	repo        schema.Repository
	bmd         []bigmapdiff.BigMapDiff
}

// NewMichelsonExtendedStorage -
func NewMichelsonExtendedStorage(impl tzip.EventImplementation, name, protocol, operationID, contract string, repo schema.Repository, bmd []bigmapdiff.BigMapDiff) (*MichelsonExtendedStorage, error) {
	parser, err := tokenbalance.GetParser(name, impl.MichelsonExtendedStorageEvent.ReturnType)
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
		repo:        repo,
		bmd:         bmd,
		contract:    contract,
	}, nil
}

// Parse -
func (mes *MichelsonExtendedStorage) Parse(response gjson.Result) []tokenbalance.TokenBalance {
	balances := make([]tokenbalance.TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balance, err := mes.parser.Parse(item)
		if err != nil {
			continue
		}
		balances = append(balances, balance)
	}
	return balances
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (mes *MichelsonExtendedStorage) Normalize(value string) gjson.Result {
	tree, err := ast.NewTypedAstFromString(value)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	if err := storage.Enrich(tree, mes.bmd, true, false); err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	// val, err = storage.EnrichEmptyPointers(metadata, val)
	// if err != nil {
	// 	logger.Error(err)
	// 	return gjson.Parse(value)
	// }
	b, err := tree.ToParameters("")
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}
	return gjson.ParseBytes(b)
}
