package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

func prepareBigMapDiffsToEnrich(bmd []bigmapdiff.BigMapDiff, skipEmpty bool) []*types.BigMapDiff {
	res := make([]*types.BigMapDiff, 0)
	for i := range bmd {
		if bmd[i].Value == nil && skipEmpty {
			continue
		}
		res = append(res, &types.BigMapDiff{
			Ptr:         bmd[i].Ptr,
			Key:         bmd[i].Key,
			Value:       bmd[i].Value,
			ID:          bmd[i].ID,
			KeyHash:     bmd[i].KeyHash,
			OperationID: bmd[i].OperationID,
			Level:       bmd[i].Level,
			Address:     bmd[i].Address,
			Network:     bmd[i].Network,
			Timestamp:   bmd[i].Timestamp,
			IndexedTime: bmd[i].IndexedTime,
			Protocol:    bmd[i].Protocol,
		})
	}
	return res
}

func getBigMapDiffModels(bmd []*types.BigMapDiff) []bigmapdiff.BigMapDiff {
	res := make([]bigmapdiff.BigMapDiff, 0)
	for i := range bmd {
		res = append(res, bigmapdiff.BigMapDiff{
			Ptr:         bmd[i].Ptr,
			Key:         bmd[i].Key,
			Value:       bmd[i].Value,
			ID:          bmd[i].ID,
			KeyHash:     bmd[i].KeyHash,
			OperationID: bmd[i].OperationID,
			Level:       bmd[i].Level,
			Address:     bmd[i].Address,
			Network:     bmd[i].Network,
			Timestamp:   bmd[i].Timestamp,
			IndexedTime: bmd[i].IndexedTime,
			Protocol:    bmd[i].Protocol,
		})
	}
	return res
}

func createBigMapAst(key, value []byte, ptr int64) (*ast.BigMap, error) {
	bigMap := ast.NewBigMap(0)
	bigMap.Ptr = &ptr

	if err := bigMap.SetKeyType(key); err != nil {
		return nil, err
	}
	if err := bigMap.SetValueType(value); err != nil {
		return nil, err
	}
	return bigMap, nil
}

func getStorage(operation operation.Operation) (*ast.TypedAst, error) {
	var s ast.Script
	if err := json.Unmarshal(operation.Script, &s); err != nil {
		return nil, err
	}
	return s.StorageType()
}
