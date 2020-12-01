package storage

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const defaultPointer = -32768

type temporaryPointerData struct {
	sourcePtr int64
	binPath   string
}

func (tpd *temporaryPointerData) updateBinPath(binPath string) {
	tpd.binPath = binPath
}

// nolint
func (tpd *temporaryPointerData) updateSourcePointer(sourcePtr int64) {
	tpd.sourcePtr = sourcePtr
}

// Babylon -
type Babylon struct {
	rpc noderpc.INode
	es  elastic.IBigMapDiff

	updates           map[int64][]elastic.Model
	temporaryPointers map[int64]*temporaryPointerData
}

// NewBabylon -
func NewBabylon(rpc noderpc.INode, es elastic.IBigMapDiff) *Babylon {
	return &Babylon{
		rpc: rpc,
		es:  es,

		updates:           make(map[int64][]elastic.Model),
		temporaryPointers: make(map[int64]*temporaryPointerData),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content gjson.Result, metadata meta.Metadata, operation models.Operation) (RichStorage, error) {
	address := content.Get("destination").String()
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	var modelUpdates []elastic.Model
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := FindBigMapPointers(metadata, result.Get("storage"))
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if modelUpdates, err = b.handleBigMapDiff(result, ptrMap, address, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          modelUpdates,
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
	storage := content.Get("script.storage")

	var bm []elastic.Model
	if result.Get("big_map_diff.#").Int() > 0 {
		ptrToBin, err := FindBigMapPointers(metadata, storage)
		if err != nil || len(ptrToBin) == 0 {
			// If pointers are not found into script`s storage we try to receive storage from node and find pointers there
			// If pointers are not found after that -> throw error
			storage, err = b.rpc.GetScriptStorageJSON(address, operation.Level)
			if err != nil {
				return RichStorage{Empty: true}, err
			}
			ptrToBin, err = FindBigMapPointers(metadata, storage)
			if err != nil {
				return RichStorage{Empty: true}, err
			}
		}

		if bm, err = b.handleBigMapDiff(result, ptrToBin, address, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		Models:          bm,
		DeffatedStorage: storage.String(),
	}, nil
}

// Enrich -
func (b *Babylon) Enrich(sStorage, sPrevStorage string, bmd []models.BigMapDiff, skipEmpty, unpack bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(sStorage), nil
	}

	storage := gjson.Parse(sStorage)
	prevStorage := gjson.Parse(sPrevStorage)

	m := map[string][]interface{}{}
	for _, bm := range bmd {
		if skipEmpty && bm.Value == "" {
			continue
		}
		elt := map[string]interface{}{
			"prim": "Elt",
		}
		args := make([]interface{}, 2)
		if unpack {
			keyBytes, err := json.Marshal(bm.Key)
			if err != nil {
				return storage, err
			}
			key, err := stringer.MichelineFromBytes(keyBytes)
			if err != nil {
				return storage, err
			}
			args[0] = key.Value()

			val, err := stringer.Micheline(gjson.Parse(bm.Value))
			if err != nil {
				return storage, err
			}
			args[1] = val.Value()
		} else {
			args[0] = bm.Key
			args[1] = gjson.Parse(bm.Value).Value()
		}

		elt["args"] = args

		var res string
		if bm.BinPath != "0" {
			binPath := strings.TrimPrefix(bm.BinPath, "0/")
			p := newmiguel.GetGJSONPath(binPath)
			jsonPath, err := findPtrJSONPath(bm.Ptr, p, storage)
			if err != nil {
				jsonPath, err = findPtrJSONPath(bm.Ptr, p, prevStorage)
				if err != nil {
					return storage, err
				}
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
				return storage, err
			}
			storage = gjson.ParseBytes(b)
		} else {
			value, err := sjson.Set(storage.String(), p, arr)
			if err != nil {
				return gjson.Result{}, err
			}
			storage = gjson.Parse(value)
		}
	}
	return storage, nil
}

func (b *Babylon) handleBigMapDiff(result gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]elastic.Model, error) {
	storageModels := make([]elastic.Model, 0)

	handlers := map[string]func(gjson.Result, map[int64]string, string, models.Operation) ([]elastic.Model, error){
		"update": b.handleBigMapDiffUpdate,
		"copy":   b.handleBigMapDiffCopy,
		"remove": b.handleBigMapDiffRemove,
		"alloc":  b.handleBigMapDiffAlloc,
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
			storageModels = append(storageModels, data...)
		}
	}
	return storageModels, nil
}

func (b *Babylon) handleBigMapDiffUpdate(item gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]elastic.Model, error) {
	ptr := item.Get("big_map").Int()

	bmd := &models.BigMapDiff{
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

	newPtr := ptr
	if ptr < 0 {
		bufPtr, err := b.getSourcePointer(ptr)
		if err != nil {
			return nil, err
		}
		newPtr = bufPtr
	}

	binPath := b.getPointerBinaryPath(ptrMap, newPtr)
	if binPath != "" {
		bmd.BinPath = binPath
		if _, ok := b.temporaryPointers[ptr]; ok {
			b.temporaryPointers[ptr].updateBinPath(binPath)
		}
	}

	b.addToUpdates(bmd, ptr)
	if ptr > -1 {
		return []elastic.Model{bmd}, nil
	}
	return nil, nil
}

func (b *Babylon) handleBigMapDiffCopy(item gjson.Result, ptrMap map[int64]string, address string, operation models.Operation) ([]elastic.Model, error) {
	sourcePtr := item.Get("source_big_map").Int()
	destinationPtr := item.Get("destination_big_map").Int()

	newUpdates := make([]elastic.Model, 0)

	if destinationPtr > -1 {
		var srcPtr int64
		if sourcePtr > -1 {
			srcPtr = sourcePtr
		} else {
			bufPtr, err := b.getSourcePointer(sourcePtr)
			if err != nil {
				return nil, err
			}
			srcPtr = bufPtr
		}
		newUpdates = append(newUpdates, b.createBigMapDiffAction("copy", address, &srcPtr, &destinationPtr, operation))
	}

	bmd, err := b.getCopyBigMapDiff(sourcePtr, address, operation.Network)
	if err != nil {
		return nil, err
	}

	if err := b.updateTemporaryPointers(sourcePtr, destinationPtr, ptrMap); err != nil {
		return nil, err
	}

	if len(bmd) == 0 {
		b.updates[destinationPtr] = newUpdates
	} else {
		binPath := b.getPointerBinaryPath(ptrMap, destinationPtr)
		for i := range bmd {
			bmd[i].BinPath = binPath
			bmd[i].ID = helpers.GenerateID()
			bmd[i].Ptr = destinationPtr
			bmd[i].Address = address
			bmd[i].Level = operation.Level
			bmd[i].IndexedTime = time.Now().UnixNano() / 1000
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationID = operation.ID
			b.addToUpdates(&bmd[i], destinationPtr)

			if destinationPtr > -1 {
				newUpdates = append(newUpdates, &bmd[i])
			}
		}
	}
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffRemove(item gjson.Result, _ map[int64]string, address string, operation models.Operation) ([]elastic.Model, error) {
	ptr := item.Get("big_map").Int()
	if ptr < 0 {
		delete(b.updates, ptr)
		return nil, nil
	}
	bmd, err := b.es.GetBigMapDiffsByPtr(address, operation.Network, ptr)
	if err != nil {
		return nil, err
	}
	newUpdates := make([]elastic.Model, len(bmd))
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
	newUpdates = append(newUpdates, b.createBigMapDiffAction("remove", address, &ptr, nil, operation))
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffAlloc(item gjson.Result, _ map[int64]string, address string, operation models.Operation) ([]elastic.Model, error) {
	ptr := item.Get("big_map").Int()
	b.updates[ptr] = []elastic.Model{}
	b.temporaryPointers[ptr] = &temporaryPointerData{
		sourcePtr: defaultPointer,
	}

	var models []elastic.Model
	if ptr > -1 {
		models = append(
			models,
			b.createBigMapDiffAction("alloc", address, &ptr, nil, operation),
		)
	}

	return models, nil
}

func (b *Babylon) getDiffsFromUpdates(ptr int64) ([]models.BigMapDiff, error) {
	updates, ok := b.updates[ptr]
	if !ok {
		return nil, errors.Errorf("[handleBigMapDiffCopy] Unknown temporary pointer: %d %v", ptr, b.updates)
	}
	bmd := make([]models.BigMapDiff, 0)
	for i := range updates {
		if item, ok := updates[i].(*models.BigMapDiff); ok {
			bmd = append(bmd, *item)
		}
	}
	return bmd, nil
}

func (b *Babylon) createBigMapDiffAction(action, address string, srcPtr, dstPtr *int64, operation models.Operation) *models.BigMapAction {
	entity := &models.BigMapAction{
		ID:          helpers.GenerateID(),
		Action:      action,
		OperationID: operation.ID,
		Level:       operation.Level,
		Address:     address,
		Network:     operation.Network,
		IndexedTime: time.Now().UnixNano() / 1000,
		Timestamp:   operation.Timestamp,
	}

	if srcPtr != nil && *srcPtr > -1 {
		entity.SourcePtr = srcPtr
	}

	if dstPtr != nil && *dstPtr > -1 {
		entity.DestinationPtr = dstPtr
	}

	return entity
}

func (b *Babylon) addToUpdates(newModel elastic.Model, ptr int64) {
	if _, ok := b.updates[ptr]; !ok {
		b.updates[ptr] = []elastic.Model{newModel}
	} else {
		b.updates[ptr] = append(b.updates[ptr], newModel)
	}
}
func (b *Babylon) getSourcePointer(ptr int64) (int64, error) {
	for ptr < 0 {
		if val, ok := b.temporaryPointers[ptr]; ok {
			if val.sourcePtr == defaultPointer {
				break
			}
			ptr = val.sourcePtr
		} else {
			return ptr, errors.Errorf("Unknown temporary pointer: %d", ptr)
		}
	}
	return ptr, nil
}

func (b *Babylon) updateTemporaryPointers(src, dst int64, ptrMap map[int64]string) error {
	binPath := b.getPointerBinaryPath(ptrMap, src)
	if binPath == "" {
		binPath = b.getPointerBinaryPath(ptrMap, dst)
	}
	if binPath != "" {
		b.temporaryPointers[dst] = &temporaryPointerData{
			sourcePtr: src,
			binPath:   binPath,
		}
	} else {
		return errors.Errorf("[updateTemporaryPointers] Invalid big map pointer: %d -> %d", src, dst)
	}

	return nil
}

func (b *Babylon) getCopyBigMapDiff(src int64, address, network string) (bmd []models.BigMapDiff, err error) {
	if src > -1 {
		bmd, err = b.es.GetBigMapDiffsByPtr(address, network, src)
		if err != nil {
			return nil, err
		}
	} else {
		bmd, err = b.getDiffsFromUpdates(src)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (b *Babylon) getPointerBinaryPath(ptrMap map[int64]string, ptr int64) string {
	binPath, ok := ptrMap[ptr]
	if !ok {
		val, ok := b.temporaryPointers[ptr]
		if !ok {
			return ""
		}
		binPath = val.binPath
	}
	return binPath
}
