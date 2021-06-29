package operations

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
)

// RichStorage -
type RichStorage struct {
	rpc    noderpc.INode
	parser storage.Parser
}

// NewRichStorage -
func NewRichStorage(bigmaps bigmap.Repository, statesRepo bigmap.StateRepository, general models.GeneralRepository, rpc noderpc.INode, protocol string) (*RichStorage, error) {
	storageParser, err := storage.MakeStorageParser(bigmaps, statesRepo, general, rpc, protocol)
	if err != nil {
		return nil, err
	}
	return &RichStorage{
		rpc:    rpc,
		parser: storageParser,
	}, nil
}

// Parse -
func (p *RichStorage) Parse(data noderpc.Operation, operation *operation.Operation) (result *parsers.Result, err error) {
	switch operation.Kind {
	case types.OperationKindTransaction:
		result, err = p.parser.ParseTransaction(data, operation)
		if err != nil {
			return nil, err
		}
	case types.OperationKindOrigination:
		result, err = p.parser.ParseOrigination(data, operation)
		if err != nil {
			return nil, err
		}
		storage, err := p.rpc.GetScriptStorageRaw(operation.Destination, operation.Level)
		if err != nil {
			return nil, err
		}
		operation.DeffatedStorage = storage
	default:
		return nil, nil
	}

	if result != nil && len(result.BigMaps) > 0 {
		storageType, err := operation.AST.StorageType()
		if err != nil {
			return nil, err
		}
		if err := storageType.SettleFromBytes(operation.DeffatedStorage); err != nil {
			return nil, err
		}
		for i := range result.BigMaps {
			if result.BigMaps[i].Name == "" {
				storage.SetBigMapName(storageType, result.BigMaps[i])
			}
			storage.Tag(result.BigMaps[i])
		}
	}

	return
}
