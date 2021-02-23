package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func getResult(op gjson.Result) (gjson.Result, error) {
	result := op.Get("metadata.operation_result")
	if !result.Exists() {
		result = op.Get("result")
		if !result.Exists() {
			return gjson.Result{}, errors.Errorf("[storage.getResult] Can not find 'result'")
		}
	}
	return result, nil
}

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
	storageJSON := operation.GetScriptSection(consts.STORAGE)
	if !storageJSON.Exists() {
		return nil, errors.New("Can't find contract`s storage section")
	}
	var tree ast.UntypedAST
	if err := json.UnmarshalFromString(storageJSON.Raw, &tree); err != nil {
		return nil, err
	}
	return tree.ToTypedAST()
}
