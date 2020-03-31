package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Babylon -
type Babylon struct {
	rpc noderpc.Pool
}

// NewBabylon -
func NewBabylon(rpc noderpc.Pool) Babylon {
	return Babylon{
		rpc: rpc,
	}
}

// ParseTransaction -
func (b Babylon) ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	address := content.Get("destination").String()
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	var bm []models.BigMapDiff
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := b.binPathToPtrMap(metadata, result.Get("storage"))
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.getBigMapDiff(result, ptrMap, address, metadata, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

// ParseOrigination -
func (b Babylon) ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	address := result.Get("originated_contracts.0").String()
	storage, err := b.rpc.GetScriptStorageJSON(address, operation.Level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	var bm []models.BigMapDiff
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrToBin, err := b.binPathToPtrMap(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.getBigMapDiff(result, ptrToBin, address, metadata, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: storage.String(),
	}, nil
}

// Enrich -
func (b Babylon) Enrich(storage string, bmd []models.BigMapDiff, skipEmpty bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(storage), nil
	}

	data := gjson.Parse(storage)
	m := map[string][]interface{}{}
	for _, bm := range bmd {
		if skipEmpty && bm.Value == "" {
			continue
		}
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 2)
		args[0] = bm.Key

		val := gjson.Parse(bm.Value)
		args[1] = val.Value()

		elt["args"] = args

		var res string
		if bm.BinPath != "0" {
			binPath := strings.TrimPrefix(bm.BinPath, "0/")
			p := newmiguel.GetGJSONPath(binPath)
			jsonPath, err := b.findPtrJSONPath(bm.Ptr, p, data)
			if err != nil {
				return data, err
			}
			res = jsonPath
		}

		if _, ok := m[res]; !ok {
			m[res] = make([]interface{}, 0)
		}
		m[res] = append(m[res], elt)
	}
	for p, arr := range m {
		if p == "" {
			b, err := json.Marshal(arr)
			if err != nil {
				return data, err
			}
			data = gjson.ParseBytes(b)
		} else {
			value, err := sjson.Set(data.String(), p, arr)
			if err != nil {
				return gjson.Result{}, err
			}
			data = gjson.Parse(value)
		}
	}
	return data, nil
}

func (b Babylon) getBigMapDiff(result gjson.Result, ptrMap map[int64]string, address string, m meta.Metadata, operation models.Operation) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)

	for _, item := range result.Get("big_map_diff").Array() {
		if item.Get("action").String() != "update" {
			continue
		}

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
			OperationID: operation.ID,
			Level:       operation.Level,
			Address:     address,
			IndexedTime: operation.IndexedTime,
			Network:     operation.Network,
		})
	}
	return bmd, nil
}

func (b Babylon) binPathToPtrMap(m meta.Metadata, storage gjson.Result) (map[int64]string, error) {
	key := make(map[int64]string)
	keyInt := storage.Get("int")

	if keyInt.Exists() {
		key[keyInt.Int()] = "0"
		return key, nil
	}

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
	var buf strings.Builder

	trimmed := strings.TrimPrefix(path, "0/")
	for _, s := range strings.Split(trimmed, "/") {
		switch s {
		case "l", "s":
			buf.WriteString("#.")
		case "k":
			buf.WriteString("#.args.0.")
		case "v":
			buf.WriteString("#.args.1.")
		case "o":
			buf.WriteString("args.0.")
		default:
			buf.WriteString("args.")
			buf.WriteString(s)
			buf.WriteString(".")
		}
	}
	buf.WriteString("int")

	ptr := storage.Get(buf.String())
	if !ptr.Exists() {
		return fmt.Errorf("Path %s is not pointer: %s", path, buf.String())
	}

	for _, p := range ptr.Array() {
		if _, ok := m[p.Int()]; ok {
			return fmt.Errorf("Pointer already exists: %d", p.Int())
		}
		m[p.Int()] = path
	}

	return nil
}

func (b Babylon) findPtrJSONPath(ptr int64, path string, data gjson.Result) (string, error) {
	val := data
	parts := strings.Split(path, ".")

	var newPath strings.Builder
	for i := range parts {
		buf := val.Get(parts[i])

		if i == len(parts)-1 {
			if buf.Get("int").Exists() && buf.Get("int").Int() == ptr {
				if newPath.Len() != 0 {
					newPath.WriteString(".")
				}
				newPath.WriteString(parts[i])
				return newPath.String(), nil
			}
		}

		if parts[i] == "#" {
			for j := 0; j < int(buf.Int()); j++ {
				var bufPath strings.Builder
				fmt.Fprintf(&bufPath, "%d", j)
				if i < len(parts)-1 {
					fmt.Fprintf(&bufPath, ".%s", strings.Join(parts[i+1:], "."))
				}
				p, err := b.findPtrJSONPath(ptr, bufPath.String(), val)
				if err != nil {
					return "", err
				}
				if p != "" {
					fmt.Fprintf(&newPath, ".%s", p)
					return newPath.String(), nil
				}
			}
		} else {
			if newPath.Len() != 0 {
				newPath.WriteString(".")
			}
			newPath.WriteString(parts[i])
			val = buf
		}
	}
	return newPath.String(), nil
}
