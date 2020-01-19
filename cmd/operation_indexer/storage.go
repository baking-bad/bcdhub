package main

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
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

func getRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, op gjson.Result, level int64, network, protocol string) (*RichStorage, error) {
	kind := op.Get("kind").String()

	result := getResult(op)
	if result == nil {
		return nil, fmt.Errorf("[getDeffatedStorageNew] Can not find 'result'")
	}

	switch kind {
	case transaction:
		address := op.Get("destination").String()
		return getTransactionRichStorage(es, rpc, result, network, protocol, address, level)
	case origination:
		return getOriginationRichStorage(es, rpc, result, network, protocol, level)
	default:
		return nil, nil
	}
}

func getTransactionRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, result *gjson.Result, network, protocol, address string, level int64) (*RichStorage, error) {
	data, err := rpc.GetScriptJSON(address, level)
	if err != nil {
		return nil, err
	}

	s := data.Get("storage")
	m, err := getMetadata(es, address, level)
	if err != nil {
		return nil, err
	}
	bm, err := getBigMapDiff(result, s, network, protocol, address, level, m)
	if err != nil {
		return nil, err
	}
	return &RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

func getOriginationRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, result *gjson.Result, network, protocol string, level int64) (*RichStorage, error) {
	switch protocol {
	case contractparser.HashBabylon:
		return getOriginationBabylonRichStorage(es, rpc, result, network, protocol, level)
	default:
		return &RichStorage{
			DeffatedStorage: result.Get("storage").String(),
		}, nil
	}
}

func getOriginationBabylonRichStorage(es *elastic.Elastic, rpc *noderpc.NodeRPC, result *gjson.Result, network, protocol string, level int64) (*RichStorage, error) {
	address := result.Get("originated_contracts.0").String()
	data, err := rpc.GetScriptJSON(address, level)
	if err != nil {
		return nil, err
	}

	s := data.Get("storage")

	m, err := getMetadata(es, address, level)
	if err != nil {
		return nil, err
	}
	bm, err := getBigMapDiff(result, s, network, protocol, address, level, m)
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

func getBigMapDiff(result *gjson.Result, storage gjson.Result, network, protocol, address string, level int64, m contractparser.Metadata) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)
	for _, item := range result.Get("big_map_diff").Array() {
		switch protocol {
		case contractparser.HashBabylon:
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
					Network: network,
					Address: address,
					Level:   level,
					Ptr:     ptr,
					BinPath: binPath,
					Key:     item.Get("key").Value(),
					KeyHash: item.Get("key_hash").String(),
					Value:   item.Get("value").String(),
				})
			}
		default:
			bmd = append(bmd, models.BigMapDiff{
				Network: network,
				Address: address,
				Level:   level,
				BinPath: "00",
				Key:     item.Get("key").Value(),
				KeyHash: item.Get("key_hash").String(),
				Value:   item.Get("value").String(),
			})
		}
	}
	return bmd, nil
}

func getBinPathToPtrMap(m contractparser.Metadata, storage gjson.Result) (map[int64]string, error) {
	key := make(map[int64]string)
	for k, v := range m {
		if v.Prim != contractparser.BIGMAP {
			continue
		}
		ptr, err := getMapPtr(storage, k)
		if err != nil {
			return nil, err
		}
		key[ptr] = k
	}
	return key, nil
}

func getMapPtr(storage gjson.Result, binPath string) (int64, error) {
	path := ""

	for _, s := range binPath[1:] {
		path += fmt.Sprintf("args.%s.", string(s))
	}

	path += "int"

	ptr := storage.Get(path)
	if !ptr.Exists() {
		return 0, fmt.Errorf("Path %s is not pointer", binPath)
	}
	return ptr.Int(), nil
}
