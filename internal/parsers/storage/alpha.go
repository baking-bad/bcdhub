package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
		DeffatedStorage: result.Storage,
	}, nil
}

// ParseOrigination -
func (a *Alpha) ParseOrigination(content noderpc.Operation, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	bmd := make([]models.Model, 0)

	var storageData struct {
		Storage ast.UntypedAST `json:"storage"`
	}

	if err := json.Unmarshal(content.Script, &storageData); err != nil {
		return RichStorage{Empty: true}, err
	}

	if err := storage.Settle(storageData.Storage); err != nil {
		return RichStorage{Empty: true}, err
	}

	pair, ok := storage.Nodes[0].(*ast.Pair)
	if ok {
		bigMap, ok := pair.Args[0].(*ast.BigMap)
		if ok {
			result := content.GetResult()
			if result == nil {
				return RichStorage{Empty: true}, nil
			}

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

				b := &bigmapdiff.BigMapDiff{
					Key:              keyBytes,
					KeyHash:          keyHash,
					Value:            valBytes,
					OperationHash:    operation.Hash,
					OperationCounter: operation.Counter,
					OperationNonce:   operation.Nonce,
					Level:            operation.Level,
					Contract:         result.Originated[0],
					Network:          operation.Network,
					Timestamp:        operation.Timestamp,
					Protocol:         operation.Protocol,
					Ptr:              -1,
				}

				bmd = append(bmd, b, b.ToState())
				return false, nil
			}); err != nil {
				return RichStorage{Empty: true}, err
			}

			if len(bmd) > 0 {
				bigMap.Data = ast.NewOrderedMap()
			}
		}
	}

	b, err := storage.ToParameters(ast.DocsFull)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	return RichStorage{
		Models:          bmd,
		DeffatedStorage: b,
	}, nil
}

func (a *Alpha) getBigMapDiff(diffs []noderpc.BigMapDiff, address string, operation operation.Operation) []models.Model {
	bmd := make([]models.Model, 0)
	for i := range diffs {
		b := &bigmapdiff.BigMapDiff{
			Key:              types.Bytes(diffs[i].Key),
			KeyHash:          diffs[i].KeyHash,
			Value:            types.Bytes(diffs[i].Value),
			OperationHash:    operation.Hash,
			OperationCounter: operation.Counter,
			OperationNonce:   operation.Nonce,
			Level:            operation.Level,
			Contract:         address,
			Network:          operation.Network,
			Timestamp:        operation.Timestamp,
			Protocol:         operation.Protocol,
			Ptr:              -1,
		}
		bmd = append(bmd, b, b.ToState())
	}
	return bmd
}
