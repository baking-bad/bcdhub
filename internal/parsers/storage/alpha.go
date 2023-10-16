package storage

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Alpha -
type Alpha struct{}

// NewAlpha -
func NewAlpha() *Alpha {
	return &Alpha{}
}

// ParseTransaction -
func (a *Alpha) ParseTransaction(ctx context.Context, content noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	result := content.GetResult()
	if result == nil {
		return nil
	}
	operation.DeffatedStorage = result.Storage

	return a.getBigMapDiff(result.BigMapDiffs, *content.Destination, operation, store)
}

// ParseOrigination -
func (a *Alpha) ParseOrigination(ctx context.Context, content noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	if content.Script == nil {
		return nil
	}
	storage, err := operation.AST.StorageType()
	if err != nil {
		return err
	}

	var storageData struct {
		Storage ast.UntypedAST `json:"storage"`
	}

	if err := json.Unmarshal(content.Script, &storageData); err != nil {
		return err
	}

	if err := storage.Settle(storageData.Storage); err != nil {
		return err
	}

	pair, ok := storage.Nodes[0].(*ast.Pair)
	if ok {
		bigMap, ok := pair.Args[0].(*ast.BigMap)
		if ok {
			result := content.GetResult()
			if result == nil {
				return nil
			}

			operation.BigMapDiffs = make([]*bigmapdiff.BigMapDiff, 0)

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
					Key:         keyBytes,
					KeyHash:     keyHash,
					Value:       valBytes,
					OperationID: operation.ID,
					Level:       operation.Level,
					Contract:    result.Originated[0],
					Timestamp:   operation.Timestamp,
					ProtocolID:  operation.ProtocolID,
					Ptr:         -1,
				}
				operation.BigMapDiffs = append(operation.BigMapDiffs, b)
				state := b.ToState()
				state.Ptr = -1
				store.AddBigMapStates(state)
				return false, nil
			}); err != nil {
				return err
			}

			if len(operation.BigMapDiffs) > 0 {
				bigMap.Data = ast.NewOrderedMap()
			}
		}
	}

	b, err := storage.ToParameters(ast.DocsFull)
	if err != nil {
		return err
	}
	operation.DeffatedStorage = b
	operation.BigMapDiffsCount = len(operation.BigMapDiffs)
	return nil
}

func (a *Alpha) getBigMapDiff(diffs []noderpc.BigMapDiff, address string, operation *operation.Operation, store parsers.Store) error {
	for i := range diffs {
		b := &bigmapdiff.BigMapDiff{
			Key:         types.Bytes(diffs[i].Key),
			KeyHash:     diffs[i].KeyHash,
			Value:       types.Bytes(diffs[i].Value),
			OperationID: operation.ID,
			Level:       operation.Level,
			Contract:    address,
			Timestamp:   operation.Timestamp,
			ProtocolID:  operation.ProtocolID,
			Ptr:         -1,
		}

		operation.BigMapDiffs = append(operation.BigMapDiffs, b)
		operation.BigMapDiffsCount = len(operation.BigMapDiffs)
		state := b.ToState()
		state.Ptr = -1
		store.AddBigMapStates(state)
	}
	return nil
}
