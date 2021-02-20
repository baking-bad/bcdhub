package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	repo bigmapdiff.Repository
	rpc  noderpc.INode

	parser storage.Parser
}

// NewRichStorage -
func NewRichStorage(repo bigmapdiff.Repository, rpc noderpc.INode, protocol string) (*RichStorage, error) {
	storageParser, err := storage.MakeStorageParser(repo, protocol, false)
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
func (p *RichStorage) Parse(data gjson.Result, operation *operation.Operation) (storage.RichStorage, error) {
	switch operation.Kind {
	case consts.Transaction:
		return p.parser.ParseTransaction(data, *operation)
	case consts.Origination:
		rs, err := p.parser.ParseOrigination(data, *operation)
		if err != nil {
			return rs, err
		}
		storage, err := p.rpc.GetScriptStorageJSON(operation.Destination, operation.Level)
		if err != nil {
			return rs, err
		}
		rs.DeffatedStorage = storage.String()
		return rs, err
	default:
		return storage.RichStorage{Empty: true}, nil
	}
}
