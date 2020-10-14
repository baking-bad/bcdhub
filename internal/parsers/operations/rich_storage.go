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
	rpc noderpc.INode
	es  elastic.IElastic

	operation *models.Operation
	metadata  *meta.ContractMetadata

	storageParser storage.Parser
}

// NewRichStorage -
func NewRichStorage(es elastic.IElastic, rpc noderpc.INode, operation *models.Operation, metadata *meta.ContractMetadata) RichStorage {
	return RichStorage{
		rpc:       rpc,
		es:        es,
		operation: operation,
		metadata:  metadata,
	}
}

// Parse -
func (p RichStorage) Parse(data gjson.Result) (storage.RichStorage, error) {
	if p.storageParser == nil {
		parser, err := contractparser.MakeStorageParser(p.rpc, p.es, p.operation.Protocol, false)
		if err != nil {
			return storage.RichStorage{Empty: true}, err
		}
		p.storageParser = parser
	}

	protoSymLink, err := meta.GetProtoSymLink(p.operation.Protocol)
	if err != nil {
		return storage.RichStorage{Empty: true}, err
	}

	m, ok := p.metadata.Storage[protoSymLink]
	if !ok {
		return storage.RichStorage{Empty: true}, errors.Errorf("Unknown metadata: %s", protoSymLink)
	}

	switch p.operation.Kind {
	case consts.Transaction:
		return p.storageParser.ParseTransaction(data, m, *p.operation)
	case consts.Origination:
		rs, err := p.storageParser.ParseOrigination(data, m, *p.operation)
		if err != nil {
			return rs, err
		}
		storage, err := p.rpc.GetScriptStorageJSON(p.operation.Destination, p.operation.Level)
		if err != nil {
			return rs, err
		}
		rs.DeffatedStorage = storage.String()
		return rs, err
	}
	return storage.RichStorage{Empty: true}, nil
}
