package operations

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
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
	storageParser, err := contractparser.MakeStorageParser(rpc, repo, protocol, false)
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
func (p *RichStorage) Parse(data gjson.Result, schema *meta.ContractSchema, operation *operation.Operation) (storage.RichStorage, error) {
	protoSymLink, err := meta.GetProtoSymLink(operation.Protocol)
	if err != nil {
		return storage.RichStorage{Empty: true}, err
	}

	m, ok := schema.Storage[protoSymLink]
	if !ok {
		return storage.RichStorage{Empty: true}, errors.Errorf("Unknown metadata: %s", protoSymLink)
	}

	switch operation.Kind {
	case consts.Transaction:
		return p.parser.ParseTransaction(data, m, *operation)
	case consts.Origination:
		rs, err := p.parser.ParseOrigination(data, m, *operation)
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
