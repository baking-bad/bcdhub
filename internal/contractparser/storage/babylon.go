package storage

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Babylon -
type Babylon struct {
	es  *elastic.Elastic
	rpc *noderpc.NodeRPC
}

// NewBabylon -
func NewBabylon(es *elastic.Elastic, rpc *noderpc.NodeRPC) Babylon {
	return Babylon{
		es:  es,
		rpc: rpc,
	}
}

// ParseTransaction -
func (b Babylon) ParseTransaction(content gjson.Result, level int64, operationID string) (RichStorage, error) {
	address := content.Get("destination").String()
	data, err := b.rpc.GetScriptJSON(address, level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	m, err := meta.GetMetadata(b.es, address, consts.Babylon, "storage", level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	ptrMap, err := b.binPathToPtrMap(m, data.Get("storage"))
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	bm, err := b.getBigMapDiff(result, ptrMap, operationID, address, level, m)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

// ParseOrigination -
func (b Babylon) ParseOrigination(content gjson.Result, level int64, operationID string) (RichStorage, error) {
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	address := result.Get("originated_contracts.0").String()
	s := content.Get("script.storage")

	m, err := meta.GetMetadata(b.es, address, consts.Babylon, "storage", level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	data, err := b.rpc.GetScriptJSON(address, level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	st := data.Get("storage")
	ptrToBin, err := b.binPathToPtrMap(m, st)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	bm, err := b.getBigMapDiff(result, ptrToBin, operationID, address, level, m)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	defStorage := s.String()
	for _, p := range ptrToBin {
		trimmed := strings.TrimPrefix(p, "0/")
		gPath := miguel.GetGJSONPath(trimmed)
		defStorage, err = sjson.Set(s.String(), gPath, []interface{}{})
		if err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: defStorage,
	}, nil
}

// Enrich -
func (b Babylon) Enrich(storage string, bmd gjson.Result) (gjson.Result, error) {
	if bmd.IsArray() && len(bmd.Array()) == 0 {
		return gjson.Parse(storage), nil
	}

	data := gjson.Parse(storage)
	m := map[string][]interface{}{}
	for _, b := range bmd.Array() {
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 1)
		val := gjson.Parse(b.Get("value").String())
		args[0] = b.Get("key").Value()

		if b.Get("value").String() != "" {
			args = append(args, val.Value())
		}

		elt["args"] = args

		binPath := strings.TrimPrefix(b.Get("bin_path").String(), "0/")
		p := miguel.GetGJSONPath(binPath)
		if _, ok := m[p]; !ok {
			m[p] = make([]interface{}, 0)
		}
		m[p] = append(m[p], elt)
	}

	for p, arr := range m {
		value, err := sjson.Set(storage, p, arr)
		if err != nil {
			return gjson.Result{}, err
		}
		data = gjson.Parse(value)
	}
	return data, nil
}

func (b Babylon) getBigMapDiff(result gjson.Result, ptrMap map[int64]string, operationID, address string, level int64, m meta.Metadata) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)

	for _, item := range result.Get("big_map_diff").Array() {
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
				Address:     address,
			})
		}
	}
	return bmd, nil
}

func (b Babylon) binPathToPtrMap(m meta.Metadata, storage gjson.Result) (map[int64]string, error) {
	key := make(map[int64]string)
	for k, v := range m {
		if v.Prim != consts.BIGMAP {
			continue
		}

		if err := b.setMapPtr(storage, k, key); err != nil {
			return nil, err
		}
	}
	return key, nil
}

func (b Babylon) setMapPtr(storage gjson.Result, path string, m map[int64]string) error {
	bufPath := ""

	trimmed := strings.TrimPrefix(path, "0/")
	for _, s := range strings.Split(trimmed, "/") {
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
