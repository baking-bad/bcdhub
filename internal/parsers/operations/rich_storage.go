package operations

import (
	"context"

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
func (p *RichStorage) Parse(ctx context.Context, data noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	switch operation.Kind {
	case types.OperationKindTransaction:
		return p.parser.ParseTransaction(ctx, data, operation, store)
	case types.OperationKindOrigination:
		parsed, err := p.parser.ParseOrigination(ctx, data, operation, store)
		if err != nil {
			return err
		}
		if parsed {
			storage, err := p.rpc.GetScriptStorageRaw(ctx, operation.Destination.Address, operation.Level)
			if err != nil {
				return err
			}
			operation.DeffatedStorage = storage
		}
		return nil
	default:
		return nil
	}
}
