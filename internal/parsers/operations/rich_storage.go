package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
)

// RichStorage -
type RichStorage struct {
	repo bigmapdiff.Repository
	rpc  noderpc.INode

	parser storage.Parser
}

// NewRichStorage -
func NewRichStorage(repo bigmapdiff.Repository, rpc noderpc.INode, protocol string) (*RichStorage, error) {
	storageParser, err := storage.MakeStorageParser(repo, rpc, protocol)
	if err != nil {
		return nil, err
	}
	return &RichStorage{
		repo:   repo,
		rpc:    rpc,
		parser: storageParser,
	}, nil
}

// Parse -
func (p *RichStorage) Parse(data noderpc.Operation, operation *operation.Operation) (*parsers.Result, error) {
	switch operation.Kind {
	case types.OperationKindTransaction:
		return p.parser.ParseTransaction(data, operation)
	case types.OperationKindOrigination:
		result, err := p.parser.ParseOrigination(data, operation)
		if err != nil {
			return nil, err
		}
		if result != nil {
			storage, err := p.rpc.GetScriptStorageRaw(operation.Destination, operation.Level)
			if err != nil {
				return nil, err
			}
			operation.DeffatedStorage = storage
		}
		return result, nil
	default:
		return nil, nil
	}
}
