package main

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// RichStorage -
type RichStorage struct {
	DeffatedStorage string
	BigMapDiffs     []models.BigMapDiff
}

func getRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, op gjson.Result, level int64, protocol, operationID string) (*RichStorage, error) {
	kind := op.Get("kind").String()
	switch kind {
	case transaction:
		return getTransactionRichStorage(es, rpc, op, protocol, operationID, level)
	case origination:
		return getOriginationRichStorage(es, rpc, op, protocol, operationID, level)
	default:
		return nil, nil
	}
}

func getTransactionRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, op gjson.Result, protocol, operationID string, level int64) (*RichStorage, error) {
	address := op.Get("destination").String()
	data, err := rpc.GetScriptJSON(address, level)
	if err != nil {
		return nil, err
	}

	s := data.Get("storage")
	m, err := getMetadata(es, address, "storage", level)
	if err != nil {
		return nil, err
	}

	result := getResult(op)
	if result == nil {
		return nil, fmt.Errorf("[getDeffatedStorageNew] Can not find 'result'")
	}

	bm, err := getBigMapDiff(result, s, protocol, operationID, level, m)
	if err != nil {
		return nil, err
	}
	return &RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

func getOriginationRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, op gjson.Result, protocol, operationID string, level int64) (*RichStorage, error) {
	switch protocol {
	case consts.HashBabylon:
		return getOriginationBabylonRichStorage(es, rpc, op, protocol, operationID, level)
	default:
		return &RichStorage{
			DeffatedStorage: op.Get("script.storage").String(),
		}, nil
	}
}

func getOriginationBabylonRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, op gjson.Result, protocol, operationID string, level int64) (*RichStorage, error) {
	result := getResult(op)
	if result == nil {
		return nil, fmt.Errorf("[getDeffatedStorageNew] Can not find 'result'")
	}

	address := result.Get("originated_contracts.0").String()
	s := op.Get("script.storage")

	m, err := getMetadata(es, address, "storage", level)
	if err != nil {
		return nil, err
	}
	bm, err := getBigMapDiff(result, s, protocol, operationID, level, m)
	if err != nil {
		return nil, err
	}

	return &RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: s.String(),
	}, nil
}

func getResult(op gjson.Result) *gjson.Result {
	result := op.Get("metadata.operation_result")
	if !result.Exists() {
		result = op.Get("result")
		if !result.Exists() {
			return nil
		}
	}
	return &result
}

func getBigMapDiff(result *gjson.Result, storage gjson.Result, protocol, operationID string, level int64, m meta.Metadata) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)
	for _, item := range result.Get("big_map_diff").Array() {
		switch protocol {
		case consts.HashBabylon:
			ptrMap, err := getBinPathToPtrMap(m, storage)
			if err != nil {
				return nil, err
			}
			if item.Get("action").String() == "update" {
				ptr := item.Get("big_map").Int()
				binPath, ok := ptrMap[ptr]
				if !ok {
					return nil, fmt.Errorf("Invalid big map pointer value: %d", ptr)
				}
				bmd = append(bmd, models.BigMapDiff{
					Ptr:         ptr,
					BinPath:     binPath,
					Key:         item.Get("key").Value(),
					KeyHash:     item.Get("key_hash").String(),
					Value:       item.Get("value").String(),
					OperationID: operationID,
					Level:       level,
				})
			}
		default:
			bmd = append(bmd, models.BigMapDiff{
				BinPath:     "0/0",
				Key:         item.Get("key").Value(),
				KeyHash:     item.Get("key_hash").String(),
				Value:       item.Get("value").String(),
				OperationID: operationID,
				Level:       level,
			})
		}
	}
	return bmd, nil
}

func getBinPathToPtrMap(m meta.Metadata, storage gjson.Result) (map[int64]string, error) {
	key := make(map[int64]string)
	for k, v := range m {
		if v.Prim != consts.BIGMAP {
			continue
		}

		if err := setMapPtr(storage, k, key); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func setMapPtr(storage gjson.Result, path string, m map[int64]string) error {
	bufPath := ""

	for _, s := range strings.Split(path, "/")[1:] {
		switch s {
		case "l", "s":
			bufPath += "#."
		case "k":
			bufPath += "#.args.0"
		case "v":
			bufPath += "#.args.1"
		case "o":
			bufPath += "args.0"
		default:
			bufPath += fmt.Sprintf("args.%s.", string(s))
		}
	}

	bufPath += "int"

	ptr := storage.Get(bufPath)
	if !ptr.Exists() {
		return fmt.Errorf("Path %s is not pointer: %s", path, bufPath)
	}

	for _, p := range ptr.Array() {
		m[p.Int()] = path
	}

	return nil
}
