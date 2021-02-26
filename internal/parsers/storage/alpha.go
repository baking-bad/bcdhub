package storage

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Alpha -
type Alpha struct{}

// NewAlpha -
func NewAlpha() *Alpha {
	return &Alpha{}
}

// ParseTransaction -
func (a *Alpha) ParseTransaction(content noderpc.Operation, operation operation.Operation) (RichStorage, error) {
	result := content.GetResult()
	if result == nil {
		return RichStorage{Empty: true}, nil
	}

	return RichStorage{
		Models:          a.getBigMapDiff(result.BigMapDiffs, *content.Destination, operation),
		DeffatedStorage: string(result.Storage),
	}, nil
}

// ParseOrigination -
func (a *Alpha) ParseOrigination(content noderpc.Operation, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	pair, ok := storage.Nodes[0].(*ast.Pair)
	if !ok {
		return RichStorage{Empty: true}, nil
	}
	bigMap, ok := pair.Args[0].(*ast.BigMap)
	if !ok {
		return RichStorage{Empty: true}, nil
	}

	result := content.GetResult()
	if result == nil {
		return RichStorage{Empty: true}, nil
	}

	var storageData struct {
		Storage ast.UntypedAST `json:"storage"`
	}

	if err := json.Unmarshal(content.Script, &storageData); err != nil {
		return RichStorage{Empty: true}, err
	}

	if err := storage.Settle(storageData.Storage); err != nil {
		return RichStorage{Empty: true}, err
	}

	bmd := make([]models.Model, 0)
	if err := bigMap.Data.Range(func(key, value ast.Comparable) (bool, error) {
		k := key.(ast.Node)
		keyHash, err := ast.BigMapKeyHashFromNode(k)
		if err != nil {
			return false, err
		}
		keyBytes, err := k.ToParameters()
		if err != nil {
			return false, err
		}

		var valBytes []byte
		if value != nil {
			v := value.(ast.Node)
			valBytes, err = v.ToParameters()
			if err != nil {
				return false, err
			}
		}
		bmd = append(bmd, &bigmapdiff.BigMapDiff{
			ID:          helpers.GenerateID(),
			Key:         keyBytes,
			KeyHash:     keyHash,
			Value:       valBytes,
			OperationID: operation.ID,
			Level:       operation.Level,
			Address:     result.Originated[0],
			IndexedTime: time.Now().UnixNano() / 1000,
			Network:     operation.Network,
			Timestamp:   operation.Timestamp,
			Protocol:    operation.Protocol,
			Ptr:         -1,
		})
		return false, nil
	}); err != nil {
		return RichStorage{Empty: true}, err
	}

	if len(bmd) > 0 {
		bigMap.Data = ast.NewOrderedMap()
	}

	b, err := storage.ToParameters(ast.DocsFull)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	return RichStorage{
		Models:          bmd,
		DeffatedStorage: string(b),
	}, nil
}

func (a *Alpha) getBigMapDiff(diffs []noderpc.BigMapDiff, address string, operation operation.Operation) []models.Model {
	bmd := make([]models.Model, 0)
	for i := range diffs {
		bmd = append(bmd, &bigmapdiff.BigMapDiff{
			ID:          helpers.GenerateID(),
			Key:         diffs[i].Key,
			KeyHash:     diffs[i].KeyHash,
			Value:       diffs[i].Value,
			OperationID: operation.ID,
			Level:       operation.Level,
			Address:     address,
			IndexedTime: time.Now().UnixNano() / 1000,
			Network:     operation.Network,
			Timestamp:   operation.Timestamp,
			Protocol:    operation.Protocol,
			Ptr:         -1,
		})
	}
	return bmd
}
