package storage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Babylon -
type Babylon struct {
	rpc noderpc.Pool
	es  *elastic.Elastic

	updates map[int64][]*models.BigMapDiff
}

// NewBabylon -
func NewBabylon(rpc noderpc.Pool, es *elastic.Elastic) *Babylon {
	return &Babylon{
		rpc: rpc,
		es:  es,

		updates: make(map[int64][]*models.BigMapDiff),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	address := content.Get("destination").String()
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	var bm []*models.BigMapDiff
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := b.binPathToPtrMap(metadata, result.Get("storage"))
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(result, ptrMap, address, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: result.Get("storage").String(),
	}, nil
}

// ParseOrigination -
func (b *Babylon) ParseOrigination(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	address := result.Get("originated_contracts.0").String()
	storage, err := b.rpc.GetScriptStorageJSON(address, operation.Level)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	var bm []*models.BigMapDiff
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrToBin, err := b.binPathToPtrMap(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(result, ptrToBin, address, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		BigMapDiffs:     bm,
		DeffatedStorage: storage.String(),
	}, nil
}

// Enrich -
func (b *Babylon) Enrich(storage string, bmd []models.BigMapDiff, skipEmpty bool) (gjson.Result, error) {
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

		val, err := stringer.Micheline(gjson.Parse(bm.Value))
		if err != nil {
			return data, err
		}
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

func (b *Babylon) handleBigMapDiff(result gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]*models.BigMapDiff, error) {
	bmd := make([]*models.BigMapDiff, 0)

	handlers := map[string]func(gjson.Result, map[int64]string, string, models.Operation) ([]*models.BigMapDiff, error){
		"update": b.handleBigMapDiffUpdate,
		"copy":   b.handleBigMapDiffCopy,
		"remove": b.handleBigMapDiffRemove,
	}

	for _, item := range result.Get("big_map_diff").Array() {
		action := item.Get("action").String()
		handler, ok := handlers[action]
		if !ok {
			continue
		}
		data, err := handler(item, ptrMap, address, operation)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			bmd = append(bmd, data...)
		}
	}
	return bmd, nil
}

func (b *Babylon) handleBigMapDiffUpdate(item gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]*models.BigMapDiff, error) {
	ptr := item.Get("big_map").Int()

	bmd := models.BigMapDiff{
		ID:          helpers.GenerateID(),
		Ptr:         ptr,
		Key:         item.Get("key").Value(),
		KeyHash:     item.Get("key_hash").String(),
		Value:       item.Get("value").String(),
		OperationID: operation.ID,
		Level:       operation.Level,
		Address:     address,
		IndexedTime: time.Now().UnixNano() / 1000,
		Network:     operation.Network,
		Timestamp:   operation.Timestamp,
		Protocol:    operation.Protocol,
	}
	if ptr >= 0 {
		binPath, ok := ptrMap[ptr]
		if !ok {
			return nil, fmt.Errorf("Invalid big map pointer value: %d", ptr)
		}
		bmd.BinPath = binPath
	}

	b.addToUpdates(&bmd, ptr)
	if ptr >= 0 {
		return []*models.BigMapDiff{&bmd}, nil
	}
	return nil, nil
}

func (b *Babylon) handleBigMapDiffCopy(item gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]*models.BigMapDiff, error) {
	sourcePtr := item.Get("source_big_map").Int()
	destinationPtr := item.Get("destination_big_map").Int()

	if sourcePtr >= 0 {
		bmd, err := b.es.GetAllBigMapDiffByPtr(address, operation.Network, sourcePtr)
		if err != nil {
			return nil, err
		}
		var binPath string
		if destinationPtr >= 0 {
			bp, ok := ptrMap[destinationPtr]
			if !ok {
				return nil, fmt.Errorf("[handleBigMapDiffCopy] Invalid big map pointer value: %d", destinationPtr)
			}
			binPath = bp
		}

		newUpdates := make([]*models.BigMapDiff, len(bmd))
		for i := range bmd {
			bmd[i].ID = helpers.GenerateID()
			bmd[i].OperationID = operation.ID
			bmd[i].Level = operation.Level
			bmd[i].IndexedTime = time.Now().UnixNano() / 1000
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].Ptr = destinationPtr
			bmd[i].Address = address
			bmd[i].BinPath = binPath
			newUpdates[i] = &bmd[i]
			b.addToUpdates(newUpdates[i], destinationPtr)
		}
		if len(bmd) == 0 {
			b.updates[destinationPtr] = []*models.BigMapDiff{}
		}
		if destinationPtr >= 0 {
			return newUpdates, nil
		}
		return nil, nil
	} else if sourcePtr < 0 {
		bmd, ok := b.updates[sourcePtr]
		if !ok {
			return nil, fmt.Errorf("[handleBigMapDiffCopy] Unknown temporary pointer: %d %v", sourcePtr, b.updates)
		}
		var binPath string
		if destinationPtr >= 0 {
			bp, ok := ptrMap[destinationPtr]
			if !ok {
				return nil, fmt.Errorf("[handleBigMapDiffCopy] Invalid big map pointer value: %d", destinationPtr)
			}
			binPath = bp
		}

		newUpdates := make([]*models.BigMapDiff, len(bmd))
		for i := range bmd {
			bmd[i].ID = helpers.GenerateID()
			bmd[i].Ptr = destinationPtr
			bmd[i].Address = address
			bmd[i].Level = operation.Level
			bmd[i].IndexedTime = time.Now().UnixNano() / 1000
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationID = operation.ID
			bmd[i].BinPath = binPath
			newUpdates[i] = bmd[i]
			b.addToUpdates(newUpdates[i], destinationPtr)
		}
		if len(bmd) == 0 {
			b.updates[destinationPtr] = []*models.BigMapDiff{}
		}
		if destinationPtr >= 0 {
			return newUpdates, nil
		}
		return nil, nil
	}

	return nil, nil
}

func (b *Babylon) handleBigMapDiffRemove(item gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]*models.BigMapDiff, error) {
	ptr := item.Get("big_map").Int()
	if ptr < 0 {
		delete(b.updates, ptr)
		return nil, nil
	}
	bmd, err := b.es.GetAllBigMapDiffByPtr(address, operation.Network, ptr)
	if err != nil {
		return nil, err
	}
	newUpdates := make([]*models.BigMapDiff, len(bmd))
	for i := range bmd {
		bmd[i].ID = helpers.GenerateID()
		bmd[i].OperationID = operation.ID
		bmd[i].Level = operation.Level
		bmd[i].IndexedTime = time.Now().UnixNano() / 1000
		bmd[i].Timestamp = operation.Timestamp
		bmd[i].Value = ""
		bmd[i].ValueStrings = []string{}
		newUpdates[i] = &bmd[i]
		b.addToUpdates(newUpdates[i], ptr)
	}
	return newUpdates, nil
}

func (b *Babylon) addToUpdates(bmd *models.BigMapDiff, ptr int64) {
	if arr, ok := b.updates[bmd.Ptr]; !ok {
		b.updates[bmd.Ptr] = []*models.BigMapDiff{bmd}
	} else {
		found := false
		for j := range arr {
			if arr[j].KeyHash != bmd.KeyHash {
				continue
			}
			found = true
			break
		}

		if !found {
			b.updates[bmd.Ptr] = append(b.updates[bmd.Ptr], bmd)
		}
	}
}

func (b *Babylon) binPathToPtrMap(m meta.Metadata, storage gjson.Result) (map[int64]string, error) {
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

func (b *Babylon) setMapPtr(storage gjson.Result, path string, m map[int64]string) error {
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

func (b *Babylon) findPtrJSONPath(ptr int64, path string, data gjson.Result) (string, error) {
	val := data
	parts := strings.Split(path, ".")

	var newPath strings.Builder
	for i := range parts {
		if parts[i] == "#" && val.IsArray() {
			for idx, item := range val.Array() {
				if i == len(parts)-1 {
					if ptr != item.Get("int").Int() {
						continue
					}
					if newPath.Len() != 0 {
						newPath.WriteString(".")
					}
					fmt.Fprintf(&newPath, "%d", idx)
					return newPath.String(), nil
				}

				p := strings.Join(parts[i+1:], ".")
				np, err := b.findPtrJSONPath(ptr, p, item)
				if err != nil {
					continue
				}
				if np != "" {
					fmt.Fprintf(&newPath, ".%d.%s", idx, strings.TrimPrefix(np, "."))
					return newPath.String(), nil
				}
			}
			return "", fmt.Errorf("Invalid path")
		}

		buf := val.Get(parts[i])
		if !buf.IsArray() && !buf.IsObject() {
			return "", fmt.Errorf("Invalid path")
		}
		if i == len(parts)-1 {
			if buf.Get("int").Exists() {
				if ptr != buf.Get("int").Int() {
					return "", fmt.Errorf("Invalid path")
				}
				if newPath.Len() != 0 {
					newPath.WriteString(".")
				}
				newPath.WriteString(parts[i])
				return newPath.String(), nil
			}
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

// SetUpdates -
func (b *Babylon) SetUpdates(temp map[int64][]*models.BigMapDiff) {
	b.updates = temp
}
