package operations

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	es  elastic.IElastic
	rpc noderpc.INode

	parser storage.Parser
}

// NewRichStorage -
func NewRichStorage(es elastic.IElastic, rpc noderpc.INode, protocol string) (*RichStorage, error) {
	storageParser, err := contractparser.MakeStorageParser(rpc, es, protocol, false)
	if err != nil {
		return nil, err
	}
	return &RichStorage{
		es:     es,
		rpc:    rpc,
		parser: storageParser,
	}, nil
}

// Parse -
func (p *RichStorage) Parse(data gjson.Result, metadata *meta.ContractMetadata, operation *models.Operation) (storage.RichStorage, error) {
	protoSymLink, err := meta.GetProtoSymLink(operation.Protocol)
	if err != nil {
		return storage.RichStorage{Empty: true}, err
	}

	m, ok := metadata.Storage[protoSymLink]
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
	}
	return storage.RichStorage{Empty: true}, nil
}
