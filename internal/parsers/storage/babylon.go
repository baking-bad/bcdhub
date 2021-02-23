package storage

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Babylon -
type Babylon struct {
	repo bigmapdiff.Repository

	ptrMap            map[int64]int64
	temporaryPointers map[int64]*ast.BigMap
}

// NewBabylon -
func NewBabylon(repo bigmapdiff.Repository) *Babylon {
	return &Babylon{
		repo: repo,

		ptrMap:            make(map[int64]int64),
		temporaryPointers: make(map[int64]*ast.BigMap),
	}
}

// ParseTransaction -
func (b *Babylon) ParseTransaction(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	address := content.Get("destination").String()
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	storageData := result.Get("storage")

	var data ast.UntypedAST
	if err := json.UnmarshalFromString(storageData.String(), &data); err != nil {
		return RichStorage{Empty: true}, err
	}
	if err := storage.Settle(data); err != nil {
		return RichStorage{Empty: true}, err
	}

	var modelUpdates []models.Model
	if result.Get("big_map_diff.#").Int() > 0 {
		if modelUpdates, err = b.handleBigMapDiff(result, storage, address, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          modelUpdates,
		DeffatedStorage: storageData.String(),
	}, nil
}

// ParseOrigination -
func (b *Babylon) ParseOrigination(content gjson.Result, operation operation.Operation) (RichStorage, error) {
	storage, err := getStorage(operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}
	result, err := getResult(content)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	address := result.Get("originated_contracts.0").String()
	storageData := content.Get("script.storage")

	var data ast.UntypedAST
	if err := json.UnmarshalFromString(storageData.String(), &data); err != nil {
		return RichStorage{Empty: true}, err
	}

	if err := storage.Settle(data); err != nil {
		return RichStorage{Empty: true}, err
	}

	bm, err := b.handleBigMapDiff(result, storage, address, operation)
	if err != nil {
		return RichStorage{Empty: true}, err
	}

	return RichStorage{
		Models:          bm,
		DeffatedStorage: storageData.String(),
	}, nil
}

func (b *Babylon) handleBigMapDiff(result gjson.Result, storage *ast.TypedAst, address string, op operation.Operation) ([]models.Model, error) {
	if result.Get("big_map_diff.#").Int() == 0 {
		return []models.Model{}, nil
	}
	storageModels := make([]models.Model, 0)

	handlers := map[string]func(gjson.Result, *ast.TypedAst, string, operation.Operation) ([]models.Model, error){
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
		data, err := handler(item, storage, address, op)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			storageModels = append(storageModels, data...)
		}
	}
	return storageModels, nil
}

func (b *Babylon) handleBigMapDiffUpdate(item gjson.Result, storage *ast.TypedAst, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := item.Get("big_map").Int()

	bmd := bigmapdiff.BigMapDiff{
		ID:          helpers.GenerateID(),
		Ptr:         ptr,
		Key:         []byte(item.Get("key").String()),
		KeyHash:     item.Get("key_hash").String(),
		Value:       []byte(item.Get("value").String()),
		OperationID: operation.ID,
		Level:       operation.Level,
		Address:     address,
		IndexedTime: time.Now().UnixNano() / 1000,
		Network:     operation.Network,
		Timestamp:   operation.Timestamp,
		Protocol:    operation.Protocol,
	}

	if err := b.addDiff(&bmd, storage, ptr); err != nil {
		return nil, err
	}
	if ptr > -1 {
		return []models.Model{&bmd}, nil
	}
	return nil, nil
}

func (b *Babylon) handleBigMapDiffCopy(item gjson.Result, storage *ast.TypedAst, address string, operation operation.Operation) ([]models.Model, error) {
	sourcePtr := item.Get("source_big_map").Int()
	destinationPtr := item.Get("destination_big_map").Int()

	newUpdates := make([]models.Model, 0)

	if destinationPtr > -1 {
		var srcPtr int64
		if sourcePtr > -1 {
			srcPtr = sourcePtr
		} else {
			bufPtr, err := b.getSourcePtr(sourcePtr)
			if err != nil {
				return nil, err
			}
			srcPtr = bufPtr
		}
		newUpdates = append(newUpdates, b.createBigMapDiffAction("copy", address, &srcPtr, &destinationPtr, operation))
	}

	b.ptrMap[destinationPtr] = sourcePtr

	bmd, err := b.getCopyBigMapDiff(sourcePtr, address, operation.Network)
	if err != nil {
		return nil, err
	}

	b.updateTemporaryPointers(storage, sourcePtr, destinationPtr)

	if len(bmd) > 0 {
		for i := range bmd {
			bmd[i].ID = helpers.GenerateID()
			bmd[i].Ptr = destinationPtr
			bmd[i].Address = address
			bmd[i].Level = operation.Level
			bmd[i].IndexedTime = time.Now().UnixNano() / 1000
			bmd[i].Timestamp = operation.Timestamp
			bmd[i].OperationID = operation.ID

			if err := b.addDiff(&bmd[i], storage, destinationPtr); err != nil {
				return nil, err
			}

			if destinationPtr > -1 {
				newUpdates = append(newUpdates, &bmd[i])
			}
		}
	}
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffRemove(item gjson.Result, _ *ast.TypedAst, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := item.Get("big_map").Int()
	if ptr < 0 {
		return nil, nil
	}
	bmd, err := b.repo.GetByPtr(address, operation.Network, ptr)
	if err != nil {
		return nil, err
	}
	newUpdates := make([]models.Model, len(bmd))
	for i := range bmd {
		bmd[i].ID = helpers.GenerateID()
		bmd[i].OperationID = operation.ID
		bmd[i].Level = operation.Level
		bmd[i].IndexedTime = time.Now().UnixNano() / 1000
		bmd[i].Timestamp = operation.Timestamp
		bmd[i].Value = nil
		bmd[i].ValueStrings = []string{}
		newUpdates[i] = &bmd[i]
	}
	newUpdates = append(newUpdates, b.createBigMapDiffAction("remove", address, &ptr, nil, operation))
	return newUpdates, nil
}

func (b *Babylon) handleBigMapDiffAlloc(item gjson.Result, _ *ast.TypedAst, address string, operation operation.Operation) ([]models.Model, error) {
	ptr := item.Get("big_map").Int()
	key := []byte(item.Get("key_type").String())
	value := []byte(item.Get("value_type").String())

	bigMap, err := createBigMapAst(key, value, ptr)
	if err != nil {
		return nil, err
	}

	b.temporaryPointers[ptr] = bigMap

	var models []models.Model
	if ptr > -1 {
		models = append(
			models,
			b.createBigMapDiffAction("alloc", address, &ptr, nil, operation),
		)
	}

	return models, nil
}

func (b *Babylon) getDiffsFromUpdates(ptr int64) ([]bigmapdiff.BigMapDiff, error) {
	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return nil, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
	}
	diffs := bigMap.GetDiffs()
	return getBigMapDiffModels(diffs), nil
}

func (b *Babylon) createBigMapDiffAction(action, address string, srcPtr, dstPtr *int64, operation operation.Operation) *bigmapaction.BigMapAction {
	entity := &bigmapaction.BigMapAction{
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

func (b *Babylon) addDiff(bmd *bigmapdiff.BigMapDiff, storage *ast.TypedAst, ptr int64) error {
	if bm, ok := b.temporaryPointers[ptr]; ok {
		return bm.EnrichBigMap(prepareBigMapDiffsToEnrich([]bigmapdiff.BigMapDiff{*bmd}, false))
	}

	bm := storage.FindBigMapByPtr()
	for p := range bm {
		if p == ptr {
			b.temporaryPointers[ptr] = bm[p]
			break
		}
	}
	bigMap, ok := b.temporaryPointers[ptr]
	if !ok {
		return errors.Errorf("Can't find big map pointer: %d", ptr)
	}

	diffs := prepareBigMapDiffsToEnrich([]bigmapdiff.BigMapDiff{*bmd}, false)
	bigMap.AddDiffs(diffs...)
	b.temporaryPointers[bmd.Ptr] = bigMap
	return nil
}

func (b *Babylon) getSourcePtr(ptr int64) (int64, error) {
	if src, ok := b.ptrMap[ptr]; ok {
		return src, nil
	}
	return ptr, errors.Wrapf(ErrUnknownTemporaryPointer, "%d", ptr)
}

func (b *Babylon) updateTemporaryPointers(storage *ast.TypedAst, src, dst int64) {
	bigMap, ok := b.temporaryPointers[src]
	if !ok {
		bm := storage.FindBigMapByPtr()
		for p := range bm {
			if p == src {
				bigMap = bm[p]
				break
			}
		}
		b.temporaryPointers[src] = bigMap
	}
	dstBigMap := ast.Copy(bigMap).(*ast.BigMap)
	dstBigMap.Ptr = &dst
	b.temporaryPointers[dst] = dstBigMap
}

func (b *Babylon) getCopyBigMapDiff(src int64, address, network string) (bmd []bigmapdiff.BigMapDiff, err error) {
	if src > -1 {
		bmd, err = b.repo.GetByPtr(address, network, src)
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
