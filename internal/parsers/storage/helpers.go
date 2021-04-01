package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

func prepareBigMapDiffsToEnrich(bmd []bigmapdiff.BigMapDiff, skipEmpty bool) []*types.BigMapDiff {
	res := make([]*types.BigMapDiff, 0)
	for i := range bmd {
		if bmd[i].Value == nil && skipEmpty {
			continue
		}
		res = append(res, &types.BigMapDiff{
			Ptr:              bmd[i].Ptr,
			Key:              bmd[i].Key,
			Value:            bmd[i].Value,
			ID:               bmd[i].ID,
			KeyHash:          bmd[i].KeyHash,
			OperationHash:    bmd[i].OperationHash,
			OperationCounter: bmd[i].OperationCounter,
			OperationNonce:   bmd[i].OperationNonce,
			Level:            bmd[i].Level,
			Address:          bmd[i].Contract,
			Network:          bmd[i].Network,
			Timestamp:        bmd[i].Timestamp,
			Protocol:         bmd[i].Protocol,
		})
	}
	return res
}

func prepareBigMapStatesToEnrich(bmd []bigmapdiff.BigMapState, skipEmpty bool) []*types.BigMapDiff {
	res := make([]*types.BigMapDiff, 0)
	for i := range bmd {
		if bmd[i].Removed && skipEmpty {
			continue
		}

		item := &types.BigMapDiff{
			Ptr:     bmd[i].Ptr,
			Key:     bmd[i].Key,
			ID:      bmd[i].ID,
			KeyHash: bmd[i].KeyHash,
			Address: bmd[i].Contract,
			Network: bmd[i].Network,
		}

		if !bmd[i].Removed {
			item.Value = bmd[i].Value
		}

		res = append(res, item)
	}
	return res
}

func getBigMapDiffModels(bmd []*types.BigMapDiff) []bigmapdiff.BigMapDiff {
	res := make([]bigmapdiff.BigMapDiff, 0)
	for i := range bmd {
		res = append(res, bigmapdiff.BigMapDiff{
			Ptr:              bmd[i].Ptr,
			Key:              bmd[i].Key,
			Value:            bmd[i].Value,
			ID:               bmd[i].ID,
			KeyHash:          bmd[i].KeyHash,
			OperationHash:    bmd[i].OperationHash,
			OperationCounter: bmd[i].OperationCounter,
			OperationNonce:   bmd[i].OperationNonce,
			Level:            bmd[i].Level,
			Contract:         bmd[i].Address,
			Network:          bmd[i].Network,
			Timestamp:        bmd[i].Timestamp,
			Protocol:         bmd[i].Protocol,
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
