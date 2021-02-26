package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
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
	storageParser, err := storage.MakeStorageParser(repo, protocol)
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
func (p *RichStorage) Parse(data noderpc.Operation, operation *operation.Operation) (storage.RichStorage, error) {
	switch operation.Kind {
	case consts.Transaction:
		return p.parser.ParseTransaction(data, *operation)
	case consts.Origination:
		rs, err := p.parser.ParseOrigination(data, *operation)
		if err != nil {
			return rs, err
		}
		storage, err := p.rpc.GetScriptStorageRaw(operation.Destination, operation.Level)
		if err != nil {
			return rs, err
		}
		rs.DeffatedStorage = string(storage)
		return rs, err
	default:
		return storage.RichStorage{Empty: true}, nil
	}
}
