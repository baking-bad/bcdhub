package storage

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
)

func prepareBigMapDiffsToEnrich(bmd []bigmap.Diff, skipEmpty bool) []*types.BigMapDiff {
	res := make([]*types.BigMapDiff, 0)
	for i := range bmd {
		if bmd[i].Value == nil && skipEmpty {
			continue
		}
		res = append(res, &types.BigMapDiff{
			Ptr:         bmd[i].BigMap.Ptr,
			Key:         bmd[i].Key,
			Value:       bmd[i].Value,
			ID:          bmd[i].ID,
			KeyHash:     bmd[i].KeyHash,
			OperationID: bmd[i].OperationID,
			Level:       bmd[i].Level,
			Address:     bmd[i].BigMap.Contract,
			Network:     bmd[i].BigMap.Network.String(),
			Timestamp:   bmd[i].Timestamp,
			Protocol:    bmd[i].ProtocolID,
		})
	}
	return res
}

func prepareBigMapStatesToEnrich(bmd []bigmap.State, skipEmpty bool) []*types.BigMapDiff {
	res := make([]*types.BigMapDiff, 0)
	for i := range bmd {
		if bmd[i].Removed && skipEmpty {
			continue
		}

		item := &types.BigMapDiff{
			Ptr:     bmd[i].BigMap.Ptr,
			Key:     bmd[i].Key,
			ID:      bmd[i].ID,
			KeyHash: bmd[i].KeyHash,
			Address: bmd[i].BigMap.Contract,
			Network: bmd[i].BigMap.Network.String(),
		}

		if !bmd[i].Removed {
			item.Value = bmd[i].Value
		}

		res = append(res, item)
	}
	return res
}

func getBigMapDiffModels(bmd []*types.BigMapDiff) []bigmap.Diff {
	res := make([]bigmap.Diff, 0)
	for i := range bmd {
		res = append(res, bigmap.Diff{
			// Ptr:         bmd[i].Ptr,
			Key:         bmd[i].Key,
			Value:       bmd[i].Value,
			ID:          bmd[i].ID,
			KeyHash:     bmd[i].KeyHash,
			OperationID: bmd[i].OperationID,
			Level:       bmd[i].Level,
			// Contract:    bmd[i].Address,
			// Network:     modelTypes.NewNetwork(bmd[i].Network),
			Timestamp:  bmd[i].Timestamp,
			ProtocolID: bmd[i].Protocol,
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

func getStrings(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var tree ast.UntypedAST
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree.GetStrings(true)
}

func setBigMapDiffsStrings(bmd *bigmap.Diff) error {
	keyStrings, err := getStrings(bmd.KeyBytes())
	if err != nil {
		return err
	}
	bmd.KeyStrings = keyStrings

	if bmd.Value != nil {
		valStrings, err := getStrings(bmd.ValueBytes())
		if err != nil {
			return err
		}
		bmd.ValueStrings = valStrings
	}
	return nil
}
